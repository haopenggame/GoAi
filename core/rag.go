package core

import (
	"context"
	"fmt"
)

// RAGChain 表示检索增强生成链
type RAGChain struct {
	retriever        DocumentRetriever
	chatModel        ChatModel
	promptTemplate   PromptTemplate
	contentFormatter ContentFormatter
}

// NewRAGChain 创建新的 RAG 链
func NewRAGChain(chatModel ChatModel, retriever DocumentRetriever) *RAGChain {
	template := `你是一个有用的助手。请使用以下上下文来回答问题。
如果你不知道答案，请直接说你不知道。

上下文：
{{.context}}

问题: {{.question}}

回答:`

	return &RAGChain{
		retriever:        retriever,
		chatModel:        chatModel,
		promptTemplate:   NewPromptTemplate(template),
		contentFormatter: DefaultContentFormatter{},
	}
}

// WithPromptTemplate 设置自定义提示词模板
func (r *RAGChain) WithPromptTemplate(template PromptTemplate) *RAGChain {
	r.promptTemplate = template
	return r
}

// WithContentFormatter 设置自定义内容格式化器
func (r *RAGChain) WithContentFormatter(formatter ContentFormatter) *RAGChain {
	r.contentFormatter = formatter
	return r
}

// Run 使用给定查询执行 RAG 链
func (r *RAGChain) Run(ctx context.Context, query string) (string, error) {
	// 检索文档
	documents, err := r.retriever.Retrieve(ctx, query, RetrieveOptions{TopK: 5})
	if err != nil {
		return "", fmt.Errorf("文档检索失败: %w", err)
	}

	// 格式化上下文
	context := r.contentFormatter.Format(documents)

	// 构建提示词
	rendered, err := r.promptTemplate.Render(map[string]interface{}{
		"context":  context,
		"question": query,
	})
	if err != nil {
		return "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	prompt := Prompt{
		Messages: []Message{
			{Role: RoleUser, Content: rendered},
		},
	}

	// 调用模型
	response, err := r.chatModel.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("模型调用失败: %w", err)
	}

	return response.Result.Output.Content, nil
}

// DocumentLoader 定义从各种来源加载文档的接口
type DocumentLoader interface {
	Load(ctx context.Context) ([]Document, error)
}

// TextLoader 从文本加载文档
type TextLoader struct {
	Text string
}

// Load 实现 DocumentLoader 接口
func (l *TextLoader) Load(ctx context.Context) ([]Document, error) {
	return []Document{NewDocument(l.Text)}, nil
}

// StringLoader 从字符串切片加载文档
type StringLoader struct {
	Texts []string
}

// Load 实现 DocumentLoader 接口
func (l *StringLoader) Load(ctx context.Context) ([]Document, error) {
	docs := make([]Document, len(l.Texts))
	for i, text := range l.Texts {
		docs[i] = NewDocument(text)
	}
	return docs, nil
}

// TextSplitter 定义文档分割的接口
type TextSplitter interface {
	Split(documents []Document) []Document
}

// RecursiveCharacterTextSplitter 按字符递归分割文本
type RecursiveCharacterTextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	Separators   []string
}

// NewRecursiveCharacterTextSplitter 创建带有默认值的递归字符文本分割器
func NewRecursiveCharacterTextSplitter() *RecursiveCharacterTextSplitter {
	return &RecursiveCharacterTextSplitter{
		ChunkSize:    1000,
		ChunkOverlap: 200,
		Separators:   []string{"\n\n", "\n", " ", ""},
	}
}

// Split 实现 TextSplitter 接口
func (s *RecursiveCharacterTextSplitter) Split(documents []Document) []Document {
	var result []Document
	for _, doc := range documents {
		chunks := s.splitText(doc.Content)
		for _, chunk := range chunks {
			newDoc := NewDocument(chunk)
			newDoc.Metadata = copyMetadata(doc.Metadata)
			result = append(result, newDoc)
		}
	}
	return result
}

func (s *RecursiveCharacterTextSplitter) splitText(text string) []string {
	var chunks []string
	if len(text) <= s.ChunkSize {
		return []string{text}
	}

	// 按块大小和重叠进行简单分割
	for i := 0; i < len(text); i += s.ChunkSize - s.ChunkOverlap {
		end := i + s.ChunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
		if end == len(text) {
			break
		}
	}
	return chunks
}

func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// ETLPipeline 表示文档的提取-转换-加载管道
type ETLPipeline struct {
	loader       DocumentLoader
	transformers []DocumentTransformer
	writer       DocumentWriter
}

// NewETLPipeline 创建新的 ETL 管道
func NewETLPipeline(loader DocumentLoader, writer DocumentWriter) *ETLPipeline {
	return &ETLPipeline{
		loader:       loader,
		transformers: make([]DocumentTransformer, 0),
		writer:       writer,
	}
}

// AddTransformer 向管道添加转换器
func (p *ETLPipeline) AddTransformer(transformer DocumentTransformer) *ETLPipeline {
	p.transformers = append(p.transformers, transformer)
	return p
}

// Run 执行 ETL 管道
func (p *ETLPipeline) Run(ctx context.Context) error {
	documents, err := p.loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("加载失败: %w", err)
	}

	for _, transformer := range p.transformers {
		documents, err = transformer.Transform(ctx, documents)
		if err != nil {
			return fmt.Errorf("转换失败: %w", err)
		}
	}

	if err := p.writer.Write(ctx, documents); err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	return nil
}

// SplitterTransformer 将 TextSplitter 包装为 DocumentTransformer
type SplitterTransformer struct {
	Splitter TextSplitter
}

// Transform 实现 DocumentTransformer 接口
func (t *SplitterTransformer) Transform(ctx context.Context, documents []Document) ([]Document, error) {
	return t.Splitter.Split(documents), nil
}

// MetadataTransformer 添加或更新文档的元数据
type MetadataTransformer struct {
	Metadata map[string]interface{}
}

// Transform 实现 DocumentTransformer 接口
func (t *MetadataTransformer) Transform(ctx context.Context, documents []Document) ([]Document, error) {
	for i := range documents {
		if documents[i].Metadata == nil {
			documents[i].Metadata = make(map[string]interface{})
		}
		for k, v := range t.Metadata {
			documents[i].Metadata[k] = v
		}
	}
	return documents, nil
}

// ContentCleaner 清理文档内容
type ContentCleaner struct {
	Cleaner func(string) string
}

// Transform 实现 DocumentTransformer 接口
func (t *ContentCleaner) Transform(ctx context.Context, documents []Document) ([]Document, error) {
	if t.Cleaner == nil {
		return documents, nil
	}
	for i := range documents {
		documents[i].Content = t.Cleaner(documents[i].Content)
	}
	return documents, nil
}
