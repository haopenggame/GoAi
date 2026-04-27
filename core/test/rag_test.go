package test

import (
	"context"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

// MockEmbeddingModel 是一个模拟的嵌入模型
type MockEmbeddingModel struct {
	embedding []float32
}

func (m *MockEmbeddingModel) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = m.embedding
	}
	return result, nil
}

func (m *MockEmbeddingModel) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	return m.embedding, nil
}

func (m *MockEmbeddingModel) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return m.embedding, nil
}

func (m *MockEmbeddingModel) Dimensions() int {
	return len(m.embedding)
}

// MockDocumentRetriever 是一个模拟的文档检索器
type MockDocumentRetriever struct {
	documents []core.Document
}

func (r *MockDocumentRetriever) Retrieve(ctx context.Context, query string, options core.RetrieveOptions) ([]core.Document, error) {
	return r.documents, nil
}

// MemoryDocumentWriter 是一个内存文档写入器
type MemoryDocumentWriter struct {
	documents []core.Document
}

func (w *MemoryDocumentWriter) Write(ctx context.Context, documents []core.Document) error {
	w.documents = documents
	return nil
}

func TestRAGChain(t *testing.T) {
	t.Run("创建 RAG 链", func(t *testing.T) {
		// 创建模拟模型和检索器
		chatModel := &MockChatModel{response: "Hello, World!"}
		retriever := &MockDocumentRetriever{
			documents: []core.Document{
				core.NewDocument("这是一个测试文档"),
			},
		}

		// 创建 RAG 链
		ragChain := core.NewRAGChain(chatModel, retriever)
		assert.NotNil(t, ragChain)
	})

	t.Run("执行 RAG 链", func(t *testing.T) {
		// 创建模拟模型和检索器
		chatModel := &MockChatModel{response: "这是一个测试回答"}
		retriever := &MockDocumentRetriever{
			documents: []core.Document{
				core.NewDocument("这是一个测试文档"),
			},
		}

		// 创建 RAG 链
		ragChain := core.NewRAGChain(chatModel, retriever)

		// 执行 RAG 链
		ctx := context.Background()
		result, err := ragChain.Run(ctx, "测试问题")
		assert.NoError(t, err)
		assert.Equal(t, "这是一个测试回答", result)
	})

	t.Run("自定义提示词模板", func(t *testing.T) {
		// 创建模拟模型和检索器
		chatModel := &MockChatModel{response: "自定义模板回答"}
		retriever := &MockDocumentRetriever{
			documents: []core.Document{
				core.NewDocument("测试文档"),
			},
		}

		// 创建自定义提示词模板
		template := core.NewPromptTemplate("自定义模板: {{.question}}")

		// 创建 RAG 链并设置自定义模板
		ragChain := core.NewRAGChain(chatModel, retriever)
		ragChain.WithPromptTemplate(template)

		// 执行 RAG 链
		ctx := context.Background()
		result, err := ragChain.Run(ctx, "测试问题")
		assert.NoError(t, err)
		assert.Equal(t, "自定义模板回答", result)
	})
}

func TestETLPipeline(t *testing.T) {
	t.Run("创建和执行 ETL 管道", func(t *testing.T) {
		// 创建文本加载器
		loader := &core.TextLoader{Text: "这是一个测试文本"}

		// 创建内存文档写入器
		writer := &MemoryDocumentWriter{}

		// 创建 ETL 管道
		pipeline := core.NewETLPipeline(loader, writer)

		// 添加文本分割器
		splitter := core.NewRecursiveCharacterTextSplitter()
		pipeline.AddTransformer(&core.SplitterTransformer{Splitter: splitter})

		// 添加元数据转换器
		pipeline.AddTransformer(&core.MetadataTransformer{
			Metadata: map[string]interface{}{"source": "test"},
		})

		// 执行 ETL 管道
		ctx := context.Background()
		err := pipeline.Run(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, writer.documents)
		assert.Equal(t, "test", writer.documents[0].Metadata["source"])
	})
}

func TestTextSplitter(t *testing.T) {
	t.Run("递归字符文本分割器", func(t *testing.T) {
		splitter := core.NewRecursiveCharacterTextSplitter()
		splitter.ChunkSize = 50
		splitter.ChunkOverlap = 10

		documents := []core.Document{
			core.NewDocument("这是一个测试文本，用于测试文本分割器的功能。这是一个较长的文本，应该被分割成多个块。"),
		}

		result := splitter.Split(documents)
		assert.Greater(t, len(result), 1)
		for _, doc := range result {
			assert.LessOrEqual(t, len(doc.Content), 50)
		}
	})
}

func TestDocumentLoaders(t *testing.T) {
	t.Run("文本加载器", func(t *testing.T) {
		loader := &core.TextLoader{Text: "测试文本"}
		documents, err := loader.Load(context.Background())
		assert.NoError(t, err)
		assert.Len(t, documents, 1)
		assert.Equal(t, "测试文本", documents[0].Content)
	})

	t.Run("字符串加载器", func(t *testing.T) {
		loader := &core.StringLoader{Texts: []string{"文本1", "文本2"}}
		documents, err := loader.Load(context.Background())
		assert.NoError(t, err)
		assert.Len(t, documents, 2)
		assert.Equal(t, "文本1", documents[0].Content)
		assert.Equal(t, "文本2", documents[1].Content)
	})
}