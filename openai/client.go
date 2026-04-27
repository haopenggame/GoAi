package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-spring/ai/core"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-3.5-turbo"
)

// Client 是 OpenAI API 客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
	debug      bool
}

// ClientOption 配置 Client 的选项函数
type ClientOption func(*Client)

// WithBaseURL 设置基础 URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient 设置 HTTP 客户端
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithModel 设置默认模型
func WithModel(model string) ClientOption {
	return func(c *Client) {
		c.model = model
	}
}

// WithDebug 启用调试模式
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// NewClient 创建新的 OpenAI 客户端
func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		model: defaultModel,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	var requestBody []byte
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		requestBody = jsonBody
		bodyReader = bytes.NewReader(jsonBody)
	}

	if c.debug {
		fmt.Printf("[DEBUG] HTTP Request: %s %s\n", method, c.baseURL+path)
		if len(requestBody) > 0 {
			fmt.Printf("[DEBUG] Request Body: %s\n", string(requestBody))
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	if c.debug {
		fmt.Printf("[DEBUG] HTTP Response: Status %d\n", resp.StatusCode)
		if len(respBody) > 0 {
			fmt.Printf("[DEBUG] Response Body: %s\n", string(respBody))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ChatCompletionRequest 表示聊天补全请求
type ChatCompletionRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Temperature      float32       `json:"temperature,omitempty"`
	TopP             float32       `json:"top_p,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	PresencePenalty  float32       `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32       `json:"frequency_penalty,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	Tools            []Tool        `json:"tools,omitempty"`
	ToolChoice       string        `json:"tool_choice,omitempty"`
}

// ChatMessage 表示聊天补全 API 中的消息
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall 表示 API 响应中的工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 表示 API 中的函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool 表示 API 中的工具定义
type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 表示 API 中的函数定义
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatCompletionResponse 表示聊天补全响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 表示聊天补全响应中的选项
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 表示 API 响应中的令牌使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest 表示嵌入请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse 表示嵌入响应
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  Usage           `json:"usage"`
}

// EmbeddingData 表示嵌入数据
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// toCoreMessages 将 API 消息转换为核心消息
func toCoreMessages(messages []ChatMessage) []core.Message {
	result := make([]core.Message, len(messages))
	for i, msg := range messages {
		result[i] = core.Message{
			Role:       core.MessageRole(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if len(msg.ToolCalls) > 0 {
			result[i].ToolCalls = make([]core.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				result[i].ToolCalls[j] = core.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: core.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}
	return result
}

// fromCoreMessages 将核心消息转换为 API 消息
func fromCoreMessages(messages []core.Message) []ChatMessage {
	result := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		result[i] = ChatMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if len(msg.ToolCalls) > 0 {
			result[i].ToolCalls = make([]ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				result[i].ToolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}
	return result
}

// fromCoreTools 将核心工具定义转换为 API 工具
func fromCoreTools(tools []core.ToolDefinition) []Tool {
	result := make([]Tool, len(tools))
	for i, tool := range tools {
		result[i] = Tool{
			Type: tool.Type,
			Function: FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return result
}
