package core

import (
	"context"
	"io"
)

// MessageRole 定义对话消息的角色类型
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message 表示对话中的单条消息
type Message struct {
	Role       MessageRole
	Content    string
	Metadata   map[string]interface{}
	ToolCalls  []ToolCall
	ToolCallID string
}

// ToolCall 表示模型请求的工具调用
type ToolCall struct {
	ID       string
	Type     string
	Function FunctionCall
}

// FunctionCall 表示函数调用的详细信息
type FunctionCall struct {
	Name      string
	Arguments string
}

// Prompt 表示包含消息和选项的提示词
type Prompt struct {
	Messages []Message
	Options  ChatOptions
}

// ChatOptions 包含聊天补全请求的选项
type ChatOptions struct {
	Model            string
	Temperature      float32
	TopP             float32
	MaxTokens        int
	PresencePenalty  float32
	FrequencyPenalty float32
	StopSequences    []string
	Tools            []ToolDefinition
	ToolChoice       string
}

// ToolDefinition 定义模型可调用的工具
type ToolDefinition struct {
	Type     string
	Function FunctionDefinition
}

// FunctionDefinition 定义函数的参数结构
type FunctionDefinition struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// ChatResponse 表示聊天模型的响应
type ChatResponse struct {
	Result   GenerationResult
	Metadata ResponseMetadata
}

// GenerationResult 表示单次生成的结果
type GenerationResult struct {
	Output   Message
	Metadata GenerationMetadata
}

// ResponseMetadata 包含响应的元数据
type ResponseMetadata struct {
	ID    string
	Model string
	Usage TokenUsage
	Raw   map[string]interface{}
}

// GenerationMetadata 包含生成的元数据
type GenerationMetadata struct {
	FinishReason string
	Index        int
	Raw          map[string]interface{}
}

// TokenUsage 包含令牌使用信息
type TokenUsage struct {
	PromptTokens     int
	GenerationTokens int
	TotalTokens      int
}

// ChatModel 定义聊天补全模型的接口
type ChatModel interface {
	Call(ctx context.Context, prompt Prompt) (ChatResponse, error)
	Stream(ctx context.Context, prompt Prompt) (ChatStream, error)
}

// ChatStream 表示流式聊天响应
type ChatStream interface {
	io.Closer
	Next() (ChatResponseChunk, error)
}

// ChatResponseChunk 表示流式响应的一个数据块
type ChatResponseChunk struct {
	Content      string
	Role         MessageRole
	ToolCalls    []ToolCall
	FinishReason string
	Metadata     map[string]interface{}
}

// EmbeddingModel 定义嵌入模型的接口
type EmbeddingModel interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	EmbedDocument(ctx context.Context, text string) ([]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	Dimensions() int
}

// EmbeddingOptions 包含嵌入请求的选项
type EmbeddingOptions struct {
	Model string
}

// ImageModel 定义图像生成模型的接口
type ImageModel interface {
	Generate(ctx context.Context, prompt string, options ImageOptions) (ImageResponse, error)
}

// ImageOptions 包含图像生成的选项
type ImageOptions struct {
	Model          string
	Width          int
	Height         int
	NumberOfImages int
	ResponseFormat string
}

// ImageResponse 表示图像模型的响应
type ImageResponse struct {
	Images   []Image
	Metadata ResponseMetadata
}

// Image 表示生成的图像
type Image struct {
	URL           string
	B64JSON       string
	RevisedPrompt string
}

// AudioModel 定义音频转录和合成的接口
type AudioModel interface {
	Transcribe(ctx context.Context, audio []byte, options AudioOptions) (string, error)
	Synthesize(ctx context.Context, text string, options AudioOptions) ([]byte, error)
}

// AudioOptions 包含音频请求的选项
type AudioOptions struct {
	Model          string
	Language       string
	Prompt         string
	ResponseFormat string
	Voice          string
	Speed          float32
}

// TaskStatus 定义异步任务状态类型
type TaskStatus string

const (
	TaskProcessing TaskStatus = "PROCESSING"
	TaskSuccess    TaskStatus = "SUCCESS"
	TaskFail       TaskStatus = "FAIL"
)

// AsyncImageModel 定义异步图像生成模型的接口
type AsyncImageModel interface {
	CreateImageTask(ctx context.Context, prompt string, options AsyncImageOptions) (AsyncTask, error)
}

// AsyncVideoModel 定义异步视频生成模型的接口
type AsyncVideoModel interface {
	CreateVideoTask(ctx context.Context, prompt string, options AsyncVideoOptions) (AsyncTask, error)
}

// AsyncTaskQuery 定义异步任务查询的接口
type AsyncTaskQuery interface {
	QueryTask(ctx context.Context, taskID string) (AsyncTaskResult, error)
	WaitForTask(ctx context.Context, taskID string, options PollOptions) (AsyncTaskResult, error)
}

// AsyncImageOptions 包含异步图像生成的选项
type AsyncImageOptions struct {
	Model            string
	Size             string
	Quality          string
	WatermarkEnabled *bool
	UserID           string
}

// AsyncVideoOptions 包含异步视频生成的选项
type AsyncVideoOptions struct {
	Model            string
	Quality          string
	WithAudio        *bool
	WatermarkEnabled *bool
	ImageURL         interface{}
	Size             string
	FPS              int
	Duration         int
	RequestID        string
	UserID           string
}

// AsyncTask 表示异步任务
type AsyncTask struct {
	ID         string
	RequestID  string
	Model      string
	TaskStatus TaskStatus
	Provider   string
	CreatedAt  int64
	Raw        map[string]interface{}
}

// AsyncTaskResult 表示异步任务查询结果
type AsyncTaskResult struct {
	Task         AsyncTask
	ImageResults []ImageResult
	VideoResults []VideoResult
}

// ImageResult 表示图像生成结果
type ImageResult struct {
	URL string
}

// VideoResult 表示视频生成结果
type VideoResult struct {
	URL           string
	CoverImageURL string
}

// PollOptions 包含轮询任务的选项
type PollOptions struct {
	Interval   int
	MaxRetries int
}
