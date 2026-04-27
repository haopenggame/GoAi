package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("需要设置 OPENAI_API_KEY 环境变量")
	}

	ctx := context.Background()

	client := openai.NewClient(apiKey)
	embeddingModel := openai.NewEmbeddingModel(client)
	vectorStore := core.NewSimpleVectorStore(embeddingModel)
	chatModel := openai.NewChatModel(client)

	fmt.Println("=== 示例 1: 基本 RAG ===")

	documents := []core.Document{
		core.NewDocument("埃菲尔铁塔位于法国巴黎，建于1889年。"),
		core.NewDocument("长城是世界七大奇迹之一。"),
		core.NewDocument("斗兽场是意大利罗马中心的椭圆形竞技场。"),
	}

	if err := vectorStore.Add(ctx, documents); err != nil {
		log.Fatalf("添加文档失败: %v", err)
	}

	retriever := &simpleRetriever{store: vectorStore}
	ragChain := core.NewRAGChain(chatModel, retriever)

	result, err := ragChain.Run(ctx, "埃菲尔铁塔在哪里？")
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("回答: %s\n\n", result)
	}

	fmt.Println("=== 示例 2: 使用 ChatClient 顾问的 RAG ===")

	qaAdvisor := core.NewQuestionAnswerAdvisor(vectorStore, core.SearchOptions{
		TopK:                2,
		SimilarityThreshold: 0.0,
	})

	chatClient := core.NewChatClient(chatModel).
		Builder().
		WithDefaultAdvisors(qaAdvisor).
		Build()

	response, err := chatClient.Prompt().
		User("告诉我关于斗兽场的信息。").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("回答: %s\n\n", response.Content())
	}

	fmt.Println("=== 示例 3: 文档处理管道 ===")

	loader := &core.StringLoader{
		Texts: []string{
			"文档1: Go是Google开发的编程语言。",
			"文档2: Python以其简洁和可读性著称。",
			"文档3: Rust提供无垃圾回收的内存安全。",
		},
	}

	splitter := core.NewRecursiveCharacterTextSplitter()
	vectorStore2 := core.NewSimpleVectorStore(embeddingModel)

	pipeline := core.NewETLPipeline(loader, &core.VectorStoreWriter{Store: vectorStore2})
	pipeline.AddTransformer(&core.SplitterTransformer{Splitter: splitter})
	pipeline.AddTransformer(&core.MetadataTransformer{
		Metadata: map[string]interface{}{
			"category": "programming",
		},
	})

	if err := pipeline.Run(ctx); err != nil {
		log.Fatalf("管道执行失败: %v", err)
	}

	fmt.Println("文档处理管道执行成功！")
}

type simpleRetriever struct {
	store core.VectorStore
}

func (r *simpleRetriever) Retrieve(ctx context.Context, query string, options core.RetrieveOptions) ([]core.Document, error) {
	return r.store.SimilaritySearch(ctx, query, core.SearchOptions{
		TopK:                options.TopK,
		SimilarityThreshold: 0.0,
		Filter:              options.Filter,
	})
}
