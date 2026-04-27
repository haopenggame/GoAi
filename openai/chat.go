package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-spring/ai/core"
)

// ChatModel 实现 core.ChatModel 的 OpenAI 聊天模型
type ChatModel struct {
	client *Client
}

// NewChatModel 创建新的 OpenAI 聊天模型
func NewChatModel(client *Client) *ChatModel {
	return &ChatModel{client: client}
}

// Call 实现 core.ChatModel 接口
func (m *ChatModel) Call(ctx context.Context, prompt core.Prompt) (core.ChatResponse, error) {
	model := m.client.model
	if prompt.Options.Model != "" {
		model = prompt.Options.Model
	}

	req := ChatCompletionRequest{
		Model:            model,
		Messages:         fromCoreMessages(prompt.Messages),
		Temperature:      prompt.Options.Temperature,
		TopP:             prompt.Options.TopP,
		MaxTokens:        prompt.Options.MaxTokens,
		PresencePenalty:  prompt.Options.PresencePenalty,
		FrequencyPenalty: prompt.Options.FrequencyPenalty,
		Stop:             prompt.Options.StopSequences,
	}

	if len(prompt.Options.Tools) > 0 {
		req.Tools = fromCoreTools(prompt.Options.Tools)
		req.ToolChoice = prompt.Options.ToolChoice
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return core.ChatResponse{}, fmt.Errorf("聊天补全失败: %w", err)
	}

	var apiResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return core.ChatResponse{}, fmt.Errorf("反序列化响应失败: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return core.ChatResponse{}, fmt.Errorf("响应中没有选项")
	}

	choice := apiResp.Choices[0]
	return core.ChatResponse{
		Result: core.GenerationResult{
			Output: core.Message{
				Role:      core.MessageRole(choice.Message.Role),
				Content:   choice.Message.Content,
				ToolCalls: toCoreToolCalls(choice.Message.ToolCalls),
			},
			Metadata: core.GenerationMetadata{
				FinishReason: choice.FinishReason,
				Index:        choice.Index,
			},
		},
		Metadata: core.ResponseMetadata{
			ID:    apiResp.ID,
			Model: apiResp.Model,
			Usage: core.TokenUsage{
				PromptTokens:     apiResp.Usage.PromptTokens,
				GenerationTokens: apiResp.Usage.CompletionTokens,
				TotalTokens:      apiResp.Usage.TotalTokens,
			},
		},
	}, nil
}

// Stream 实现 core.ChatModel 接口
func (m *ChatModel) Stream(ctx context.Context, prompt core.Prompt) (core.ChatStream, error) {
	model := m.client.model
	if prompt.Options.Model != "" {
		model = prompt.Options.Model
	}

	req := ChatCompletionRequest{
		Model:            model,
		Messages:         fromCoreMessages(prompt.Messages),
		Temperature:      prompt.Options.Temperature,
		TopP:             prompt.Options.TopP,
		MaxTokens:        prompt.Options.MaxTokens,
		PresencePenalty:  prompt.Options.PresencePenalty,
		FrequencyPenalty: prompt.Options.FrequencyPenalty,
		Stop:             prompt.Options.StopSequences,
		Stream:           true,
	}

	if len(prompt.Options.Tools) > 0 {
		req.Tools = fromCoreTools(prompt.Options.Tools)
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.client.baseURL+"/chat/completions", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("创建流式请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.client.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := m.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("执行流式请求失败: %w", err)
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("流式请求失败，状态码 %d", resp.StatusCode)
	}

	return &chatStream{
		reader: bufio.NewReader(resp.Body),
		closer: resp.Body,
	}, nil
}

// chatStream 实现 core.ChatStream 的 OpenAI 流式聊天
type chatStream struct {
	reader *bufio.Reader
	closer io.Closer
	buffer strings.Builder
}

// Close 实现 core.ChatStream 接口
func (s *chatStream) Close() error {
	return s.closer.Close()
}

// Next 实现 core.ChatStream 接口
func (s *chatStream) Next() (core.ChatResponseChunk, error) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return core.ChatResponseChunk{}, io.EOF
			}
			return core.ChatResponseChunk{}, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return core.ChatResponseChunk{}, io.EOF
		}

		var streamResp struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		choice := streamResp.Choices[0]
		return core.ChatResponseChunk{
			Content:      choice.Delta.Content,
			Role:         core.MessageRole(choice.Delta.Role),
			FinishReason: choice.FinishReason,
			Metadata: map[string]interface{}{
				"id":    streamResp.ID,
				"model": streamResp.Model,
			},
		}, nil
	}
}

func toCoreToolCalls(toolCalls []ToolCall) []core.ToolCall {
	result := make([]core.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = core.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: core.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}
