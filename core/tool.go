package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolCallback 定义工具执行的回调函数
type ToolCallback func(ctx context.Context, arguments string) (string, error)

// ToolRegistry 管理工具定义和回调
type ToolRegistry struct {
	definitions map[string]ToolDefinition
	callbacks   map[string]ToolCallback
}

// NewToolRegistry 创建新的工具注册中心
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		definitions: make(map[string]ToolDefinition),
		callbacks:   make(map[string]ToolCallback),
	}
}

// Register 注册工具及其定义和回调
func (r *ToolRegistry) Register(definition ToolDefinition, callback ToolCallback) error {
	name := definition.Function.Name
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}
	r.definitions[name] = definition
	r.callbacks[name] = callback
	return nil
}

// GetDefinition 按名称获取工具定义
func (r *ToolRegistry) GetDefinition(name string) (ToolDefinition, bool) {
	def, ok := r.definitions[name]
	return def, ok
}

// GetCallback 按名称获取工具回调
func (r *ToolRegistry) GetCallback(name string) (ToolCallback, bool) {
	cb, ok := r.callbacks[name]
	return cb, ok
}

// GetAllDefinitions 返回所有已注册的工具定义
func (r *ToolRegistry) GetAllDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		defs = append(defs, def)
	}
	return defs
}

// Execute 执行工具调用
func (r *ToolRegistry) Execute(ctx context.Context, toolCall ToolCall) (string, error) {
	callback, ok := r.callbacks[toolCall.Function.Name]
	if !ok {
		return "", fmt.Errorf("未找到工具: %s", toolCall.Function.Name)
	}
	return callback(ctx, toolCall.Function.Arguments)
}

// ToolCallingChatModel 包装 ChatModel 以支持工具调用
type ToolCallingChatModel struct {
	model    ChatModel
	registry *ToolRegistry
}

// NewToolCallingChatModel 创建新的工具调用聊天模型
func NewToolCallingChatModel(model ChatModel, registry *ToolRegistry) *ToolCallingChatModel {
	return &ToolCallingChatModel{
		model:    model,
		registry: registry,
	}
}

// Call 实现 ChatModel 接口，支持工具调用
func (m *ToolCallingChatModel) Call(ctx context.Context, prompt Prompt) (ChatResponse, error) {
	// 如果注册中心已设置，添加工具到选项中
	if m.registry != nil {
		prompt.Options.Tools = m.registry.GetAllDefinitions()
	}

	response, err := m.model.Call(ctx, prompt)
	if err != nil {
		return ChatResponse{}, err
	}

	// 检查工具调用
	toolCalls := response.Result.Output.ToolCalls
	if len(toolCalls) == 0 {
		return response, nil
	}

	// 执行工具调用并继续对话
	messages := append(prompt.Messages, response.Result.Output)

	for _, toolCall := range toolCalls {
		result, err := m.registry.Execute(ctx, toolCall)
		if err != nil {
			result = fmt.Sprintf("错误: %v", err)
		}
		messages = append(messages, Message{
			Role:       RoleTool,
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	// 使用工具结果继续
	followUpPrompt := Prompt{
		Messages: messages,
		Options:  prompt.Options,
	}

	return m.model.Call(ctx, followUpPrompt)
}

// Stream 实现 ChatModel 接口
func (m *ToolCallingChatModel) Stream(ctx context.Context, prompt Prompt) (ChatStream, error) {
	if m.registry != nil {
		prompt.Options.Tools = m.registry.GetAllDefinitions()
	}
	return m.model.Stream(ctx, prompt)
}

// FunctionToolBuilder 帮助构建函数工具定义
type FunctionToolBuilder struct {
	name        string
	description string
	parameters  map[string]interface{}
	required    []string
}

// NewFunctionToolBuilder 创建新的函数工具构建器
func NewFunctionToolBuilder(name string) *FunctionToolBuilder {
	return &FunctionToolBuilder{
		name:       name,
		parameters: make(map[string]interface{}),
		required:   make([]string, 0),
	}
}

// WithDescription 设置描述
func (b *FunctionToolBuilder) WithDescription(description string) *FunctionToolBuilder {
	b.description = description
	return b
}

// WithParameter 添加参数
func (b *FunctionToolBuilder) WithParameter(name string, paramType string, description string, required bool) *FunctionToolBuilder {
	if b.parameters["type"] == nil {
		b.parameters["type"] = "object"
		b.parameters["properties"] = make(map[string]interface{})
	}

	properties := b.parameters["properties"].(map[string]interface{})
	properties[name] = map[string]interface{}{
		"type":        paramType,
		"description": description,
	}

	if required {
		b.required = append(b.required, name)
	}

	return b
}

// Build 构建 ToolDefinition
func (b *FunctionToolBuilder) Build() ToolDefinition {
	b.parameters["required"] = b.required
	return ToolDefinition{
		Type: "function",
		Function: FunctionDefinition{
			Name:        b.name,
			Description: b.description,
			Parameters:  b.parameters,
		},
	}
}

// ParseArguments 将 JSON 参数解析到结构体
func ParseArguments(arguments string, v interface{}) error {
	return json.Unmarshal([]byte(arguments), v)
}
