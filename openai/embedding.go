package openai

import (
	"context"
	"encoding/json"
	"fmt"
)

// EmbeddingModel 实现 core.EmbeddingModel 的 OpenAI 嵌入模型
type EmbeddingModel struct {
	client *Client
	model  string
}

// NewEmbeddingModel 创建新的 OpenAI 嵌入模型
func NewEmbeddingModel(client *Client) *EmbeddingModel {
	return &EmbeddingModel{
		client: client,
		model:  "text-embedding-ada-002",
	}
}

// WithEmbeddingModel 设置嵌入模型
func (m *EmbeddingModel) WithEmbeddingModel(model string) *EmbeddingModel {
	m.model = model
	return m
}

// Embed 实现 core.EmbeddingModel 接口
func (m *EmbeddingModel) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	req := EmbeddingRequest{
		Model: m.model,
		Input: texts,
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/embeddings", req)
	if err != nil {
		return nil, fmt.Errorf("嵌入请求失败: %w", err)
	}

	var apiResp EmbeddingResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %w", err)
	}

	result := make([][]float32, len(apiResp.Data))
	for _, data := range apiResp.Data {
		if data.Index >= 0 && data.Index < len(result) {
			result[data.Index] = data.Embedding
		}
	}

	return result, nil
}

// EmbedDocument 实现 core.EmbeddingModel 接口
func (m *EmbeddingModel) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := m.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("未返回嵌入结果")
	}
	return embeddings[0], nil
}

// EmbedQuery 实现 core.EmbeddingModel 接口
func (m *EmbeddingModel) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return m.EmbedDocument(ctx, text)
}

// Dimensions 实现 core.EmbeddingModel 接口
func (m *EmbeddingModel) Dimensions() int {
	switch m.model {
	case "text-embedding-ada-002":
		return 1536
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	default:
		return 1536
	}
}
