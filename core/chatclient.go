package core

import (
	"context"
	"fmt"
	"strings"
)

// ChatClient 提供与聊天模型交互的流式API
type ChatClient struct {
	model            ChatModel
	defaultSystem    string
	defaultOptions   ChatOptions
	defaultAdvisors  []Advisor
	templateRenderer TemplateRenderer
}

// ChatClientBuilder 用于构建 ChatClient
type ChatClientBuilder struct {
	client *ChatClient
}

// NewChatClient 创建新的 ChatClient
func NewChatClient(model ChatModel) *ChatClient {
	return &ChatClient{
		model:            model,
		defaultAdvisors:  make([]Advisor, 0),
		templateRenderer: DefaultTemplateRenderer{},
	}
}

// Builder 创建 ChatClient 的构建器
func (c *ChatClient) Builder() *ChatClientBuilder {
	return &ChatClientBuilder{
		client: &ChatClient{
			model:            c.model,
			defaultSystem:    c.defaultSystem,
			defaultOptions:   c.defaultOptions,
			defaultAdvisors:  append([]Advisor{}, c.defaultAdvisors...),
			templateRenderer: c.templateRenderer,
		},
	}
}

// WithDefaultSystem 设置默认系统消息
func (b *ChatClientBuilder) WithDefaultSystem(system string) *ChatClientBuilder {
	b.client.defaultSystem = system
	return b
}

// WithDefaultOptions 设置默认选项
func (b *ChatClientBuilder) WithDefaultOptions(options ChatOptions) *ChatClientBuilder {
	b.client.defaultOptions = options
	return b
}

// WithDefaultAdvisors 设置默认顾问
func (b *ChatClientBuilder) WithDefaultAdvisors(advisors ...Advisor) *ChatClientBuilder {
	b.client.defaultAdvisors = append(b.client.defaultAdvisors, advisors...)
	return b
}

// WithTemplateRenderer 设置模板渲染器
func (b *ChatClientBuilder) WithTemplateRenderer(renderer TemplateRenderer) *ChatClientBuilder {
	b.client.templateRenderer = renderer
	return b
}

// Build 构建 ChatClient
func (b *ChatClientBuilder) Build() *ChatClient {
	return b.client
}

// Prompt 开始构建此客户端的新提示词
func (c *ChatClient) Prompt() *ChatClientPromptBuilder {
	return &ChatClientPromptBuilder{
		client:   c,
		messages: make([]Message, 0),
		options:  c.defaultOptions,
	}
}

// ChatClientPromptBuilder 为 ChatClient 构建提示词
type ChatClientPromptBuilder struct {
	client    *ChatClient
	messages  []Message
	options   ChatOptions
	advisors  []Advisor
	variables map[string]interface{}
}

// System 添加系统消息
func (b *ChatClientPromptBuilder) System(content string) *ChatClientPromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleSystem, Content: content})
	return b
}

// User 添加用户消息
func (b *ChatClientPromptBuilder) User(content string) *ChatClientPromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleUser, Content: content})
	return b
}

// Assistant 添加助手消息
func (b *ChatClientPromptBuilder) Assistant(content string) *ChatClientPromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleAssistant, Content: content})
	return b
}

// Advisors 添加顾问
func (b *ChatClientPromptBuilder) Advisors(advisors ...Advisor) *ChatClientPromptBuilder {
	b.advisors = append(b.advisors, advisors...)
	return b
}

// WithVariable 设置模板变量
func (b *ChatClientPromptBuilder) WithVariable(key string, value interface{}) *ChatClientPromptBuilder {
	if b.variables == nil {
		b.variables = make(map[string]interface{})
	}
	b.variables[key] = value
	return b
}

// WithVariables 设置多个模板变量
func (b *ChatClientPromptBuilder) WithVariables(variables map[string]interface{}) *ChatClientPromptBuilder {
	if b.variables == nil {
		b.variables = make(map[string]interface{})
	}
	for k, v := range variables {
		b.variables[k] = v
	}
	return b
}

// WithOptions 设置聊天选项
func (b *ChatClientPromptBuilder) WithOptions(options ChatOptions) *ChatClientPromptBuilder {
	b.options = options
	return b
}

// WithModel 设置模型
func (b *ChatClientPromptBuilder) WithModel(model string) *ChatClientPromptBuilder {
	b.options.Model = model
	return b
}

// WithTemperature 设置温度参数
func (b *ChatClientPromptBuilder) WithTemperature(temperature float32) *ChatClientPromptBuilder {
	b.options.Temperature = temperature
	return b
}

// WithMaxTokens 设置最大令牌数
func (b *ChatClientPromptBuilder) WithMaxTokens(maxTokens int) *ChatClientPromptBuilder {
	b.options.MaxTokens = maxTokens
	return b
}

// WithTools 设置工具列表
func (b *ChatClientPromptBuilder) WithTools(tools []ToolDefinition) *ChatClientPromptBuilder {
	b.options.Tools = tools
	return b
}

// buildPrompt 构建最终的提示词
func (b *ChatClientPromptBuilder) buildPrompt() (Prompt, error) {
	messages := make([]Message, 0)

	// 添加默认系统消息
	if b.client.defaultSystem != "" {
		content := b.client.defaultSystem
		if b.variables != nil {
			rendered, err := b.client.templateRenderer.Render(content, b.variables)
			if err != nil {
				return Prompt{}, fmt.Errorf("渲染系统模板失败: %w", err)
			}
			content = rendered
		}
		messages = append(messages, Message{Role: RoleSystem, Content: content})
	}

	// 添加显式系统消息
	for _, msg := range b.messages {
		if msg.Role == RoleSystem {
			messages = append(messages, msg)
		}
	}

	// 添加用户和助手消息
	for _, msg := range b.messages {
		if msg.Role != RoleSystem {
			if b.variables != nil {
				rendered, err := b.client.templateRenderer.Render(msg.Content, b.variables)
				if err != nil {
					return Prompt{}, fmt.Errorf("渲染消息模板失败: %w", err)
				}
				msg.Content = rendered
			}
			messages = append(messages, msg)
		}
	}

	return Prompt{
		Messages: messages,
		Options:  b.options,
	}, nil
}

// Call 执行聊天调用并返回响应
func (b *ChatClientPromptBuilder) Call() (*ChatClientResponse, error) {
	ctx := context.Background()
	return b.CallWithContext(ctx)
}

// CallWithContext 使用上下文执行聊天调用
func (b *ChatClientPromptBuilder) CallWithContext(ctx context.Context) (*ChatClientResponse, error) {
	prompt, err := b.buildPrompt()
	if err != nil {
		return nil, err
	}

	// 执行前置顾问
	advisors := append(b.client.defaultAdvisors, b.advisors...)
	advisorContext := make(map[string]interface{})
	for _, advisor := range advisors {
		prompt, err = advisor.Before(ctx, prompt, advisorContext)
		if err != nil {
			return nil, fmt.Errorf("顾问前置处理失败: %w", err)
		}
	}

	response, err := b.client.model.Call(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 执行后置顾问（逆序）
	for i := len(advisors) - 1; i >= 0; i-- {
		response, err = advisors[i].After(ctx, response, advisorContext)
		if err != nil {
			return nil, fmt.Errorf("顾问后置处理失败: %w", err)
		}
	}

	return &ChatClientResponse{
		response: response,
		metadata: response.Metadata,
	}, nil
}

// Stream 执行流式聊天调用
func (b *ChatClientPromptBuilder) Stream() (ChatStream, error) {
	ctx := context.Background()
	return b.StreamWithContext(ctx)
}

// StreamWithContext 使用上下文执行流式聊天调用
func (b *ChatClientPromptBuilder) StreamWithContext(ctx context.Context) (ChatStream, error) {
	prompt, err := b.buildPrompt()
	if err != nil {
		return nil, err
	}
	return b.client.model.Stream(ctx, prompt)
}

// ChatClientResponse 包装聊天响应并提供便捷方法
type ChatClientResponse struct {
	response ChatResponse
	metadata ResponseMetadata
}

// Content 返回响应内容
func (r *ChatClientResponse) Content() string {
	if r == nil {
		return ""
	}
	return r.response.Result.Output.Content
}

// Message 返回输出消息
func (r *ChatClientResponse) Message() Message {
	return r.response.Result.Output
}

// Metadata 返回响应元数据
func (r *ChatClientResponse) Metadata() ResponseMetadata {
	return r.metadata
}

// ToolCalls 返回响应中的工具调用
func (r *ChatClientResponse) ToolCalls() []ToolCall {
	return r.response.Result.Output.ToolCalls
}

// Advisor 定义可以拦截和修改请求/响应的顾问接口
type Advisor interface {
	Before(ctx context.Context, prompt Prompt, advisorContext map[string]interface{}) (Prompt, error)
	After(ctx context.Context, response ChatResponse, advisorContext map[string]interface{}) (ChatResponse, error)
}

// QuestionAnswerAdvisor 提供 RAG 问答能力
type QuestionAnswerAdvisor struct {
	vectorStore    VectorStore
	searchOptions  SearchOptions
	promptTemplate PromptTemplate
}

// NewQuestionAnswerAdvisor 创建新的问答顾问
func NewQuestionAnswerAdvisor(vectorStore VectorStore, searchOptions SearchOptions) *QuestionAnswerAdvisor {
	template := `以下是上下文信息：
---------------------
{{.context}}
---------------------
根据上下文信息，不依赖先验知识，回答查询。
查询: {{.query}}`

	return &QuestionAnswerAdvisor{
		vectorStore:    vectorStore,
		searchOptions:  searchOptions,
		promptTemplate: NewPromptTemplate(template),
	}
}

// Before 实现 Advisor 接口
func (a *QuestionAnswerAdvisor) Before(ctx context.Context, prompt Prompt, advisorContext map[string]interface{}) (Prompt, error) {
	if len(prompt.Messages) == 0 {
		return prompt, nil
	}

	// 获取最后的用户消息作为查询
	var query string
	for i := len(prompt.Messages) - 1; i >= 0; i-- {
		if prompt.Messages[i].Role == RoleUser {
			query = prompt.Messages[i].Content
			break
		}
	}

	if query == "" {
		return prompt, nil
	}

	documents, err := a.vectorStore.SimilaritySearch(ctx, query, a.searchOptions)
	if err != nil {
		return prompt, fmt.Errorf("相似性搜索失败: %w", err)
	}

	if len(documents) == 0 {
		return prompt, nil
	}

	// 格式化上下文
	var contextParts []string
	for _, doc := range documents {
		contextParts = append(contextParts, doc.Content)
	}
	context := strings.Join(contextParts, "\n\n")

	// 将检索到的文档存储在顾问上下文中
	advisorContext["retrieved_documents"] = documents

	// 用上下文增强用户消息
	rendered, err := a.promptTemplate.Render(map[string]interface{}{
		"context": context,
		"query":   query,
	})
	if err != nil {
		return prompt, fmt.Errorf("渲染问答模板失败: %w", err)
	}

	// 替换最后的用户消息为增强内容
	for i := len(prompt.Messages) - 1; i >= 0; i-- {
		if prompt.Messages[i].Role == RoleUser {
			prompt.Messages[i].Content = rendered
			break
		}
	}

	return prompt, nil
}

// After 实现 Advisor 接口
func (a *QuestionAnswerAdvisor) After(ctx context.Context, response ChatResponse, advisorContext map[string]interface{}) (ChatResponse, error) {
	return response, nil
}

// SimpleLoggerAdvisor 简单的日志顾问
type SimpleLoggerAdvisor struct {
	Logger func(format string, args ...interface{})
}

// Before 实现 Advisor 接口
func (a *SimpleLoggerAdvisor) Before(ctx context.Context, prompt Prompt, advisorContext map[string]interface{}) (Prompt, error) {
	if a.Logger != nil {
		a.Logger("聊天请求: %d 条消息", len(prompt.Messages))
	}
	return prompt, nil
}

// After 实现 Advisor 接口
func (a *SimpleLoggerAdvisor) After(ctx context.Context, response ChatResponse, advisorContext map[string]interface{}) (ChatResponse, error) {
	if a.Logger != nil {
		a.Logger("聊天响应: %s", response.Result.Output.Content)
	}
	return response, nil
}
