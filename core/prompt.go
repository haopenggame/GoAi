package core

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// PromptTemplate 定义生成提示词的模板
type PromptTemplate struct {
	Template string
	Renderer TemplateRenderer
}

// TemplateRenderer 定义模板渲染的接口
type TemplateRenderer interface {
	Render(template string, variables map[string]interface{}) (string, error)
}

// DefaultTemplateRenderer 是使用 Go text/template 的默认模板渲染器
type DefaultTemplateRenderer struct{}

// Render 实现 TemplateRenderer 接口
func (r DefaultTemplateRenderer) Render(tmpl string, variables map[string]interface{}) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}
	return buf.String(), nil
}

// NewPromptTemplate 使用默认渲染器创建新的提示词模板
func NewPromptTemplate(tmpl string) PromptTemplate {
	return PromptTemplate{
		Template: tmpl,
		Renderer: DefaultTemplateRenderer{},
	}
}

// Render 使用给定变量渲染模板
func (p PromptTemplate) Render(variables map[string]interface{}) (string, error) {
	renderer := p.Renderer
	if renderer == nil {
		renderer = DefaultTemplateRenderer{}
	}
	return renderer.Render(p.Template, variables)
}

// CreatePrompt 从模板创建包含用户消息的提示词
func (p PromptTemplate) CreatePrompt(variables map[string]interface{}) (Prompt, error) {
	content, err := p.Render(variables)
	if err != nil {
		return Prompt{}, err
	}
	return Prompt{
		Messages: []Message{
			{Role: RoleUser, Content: content},
		},
	}, nil
}

// SystemPromptTemplate 是用于系统消息的提示词模板
type SystemPromptTemplate struct {
	PromptTemplate
}

// CreateSystemMessage 从模板创建系统消息
func (p SystemPromptTemplate) CreateSystemMessage(variables map[string]interface{}) (Message, error) {
	content, err := p.Render(variables)
	if err != nil {
		return Message{}, err
	}
	return Message{Role: RoleSystem, Content: content}, nil
}

// PromptBuilder 用于流式构建提示词
type PromptBuilder struct {
	messages []Message
	options  ChatOptions
}

// NewPromptBuilder 创建新的提示词构建器
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		messages: make([]Message, 0),
		options:  ChatOptions{},
	}
}

// System 添加系统消息
func (b *PromptBuilder) System(content string) *PromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleSystem, Content: content})
	return b
}

// User 添加用户消息
func (b *PromptBuilder) User(content string) *PromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleUser, Content: content})
	return b
}

// Assistant 添加助手消息
func (b *PromptBuilder) Assistant(content string) *PromptBuilder {
	b.messages = append(b.messages, Message{Role: RoleAssistant, Content: content})
	return b
}

// ToolResult 添加工具结果消息
func (b *PromptBuilder) ToolResult(toolCallID string, content string) *PromptBuilder {
	b.messages = append(b.messages, Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
	})
	return b
}

// WithOptions 设置聊天选项
func (b *PromptBuilder) WithOptions(options ChatOptions) *PromptBuilder {
	b.options = options
	return b
}

// WithModel 设置模型
func (b *PromptBuilder) WithModel(model string) *PromptBuilder {
	b.options.Model = model
	return b
}

// WithTemperature 设置温度参数
func (b *PromptBuilder) WithTemperature(temperature float32) *PromptBuilder {
	b.options.Temperature = temperature
	return b
}

// WithMaxTokens 设置最大令牌数
func (b *PromptBuilder) WithMaxTokens(maxTokens int) *PromptBuilder {
	b.options.MaxTokens = maxTokens
	return b
}

// WithTools 设置工具列表
func (b *PromptBuilder) WithTools(tools []ToolDefinition) *PromptBuilder {
	b.options.Tools = tools
	return b
}

// Build 构建最终的提示词
func (b *PromptBuilder) Build() Prompt {
	return Prompt{
		Messages: b.messages,
		Options:  b.options,
	}
}

// StringPromptTemplate 是基于简单字符串替换的提示词模板
type StringPromptTemplate struct {
	Template string
}

// NewStringPromptTemplate 创建新的字符串提示词模板
func NewStringPromptTemplate(tmpl string) StringPromptTemplate {
	return StringPromptTemplate{Template: tmpl}
}

// Render 使用变量替换渲染模板
func (p StringPromptTemplate) Render(variables map[string]interface{}) (string, error) {
	result := p.Template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result, nil
}
