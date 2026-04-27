package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-spring/ai/core"
)

// ImageModel 实现 core.ImageModel 的 OpenAI 图像模型
type ImageModel struct {
	client *Client
}

// NewImageModel 创建新的 OpenAI 图像模型
func NewImageModel(client *Client) *ImageModel {
	return &ImageModel{client: client}
}

// ImageGenerationRequest 表示图像生成请求
type ImageGenerationRequest struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

// ImageGenerationResponse 表示图像生成响应
type ImageGenerationResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// ImageData 表示响应中的图像数据
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// Generate 实现 core.ImageModel 接口
func (m *ImageModel) Generate(ctx context.Context, prompt string, options core.ImageOptions) (core.ImageResponse, error) {
	model := options.Model
	if model == "" {
		model = "dall-e-3"
	}

	size := fmt.Sprintf("%dx%d", options.Width, options.Height)
	if options.Width == 0 || options.Height == 0 {
		size = "1024x1024"
	}

	n := options.NumberOfImages
	if n <= 0 {
		n = 1
	}

	req := ImageGenerationRequest{
		Model:          model,
		Prompt:         prompt,
		N:              n,
		Size:           size,
		ResponseFormat: options.ResponseFormat,
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/images/generations", req)
	if err != nil {
		return core.ImageResponse{}, fmt.Errorf("图像生成失败: %w", err)
	}

	var apiResp ImageGenerationResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return core.ImageResponse{}, fmt.Errorf("反序列化响应失败: %w", err)
	}

	images := make([]core.Image, len(apiResp.Data))
	for i, data := range apiResp.Data {
		images[i] = core.Image{
			URL:           data.URL,
			B64JSON:       data.B64JSON,
			RevisedPrompt: data.RevisedPrompt,
		}
	}

	return core.ImageResponse{
		Images: images,
		Metadata: core.ResponseMetadata{
			Model: model,
		},
	}, nil
}
