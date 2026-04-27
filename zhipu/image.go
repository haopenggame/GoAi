package zhipu

import (
	"context"
	"fmt"

	"github.com/go-spring/ai/core"
)

type ImageModel struct {
	client *Client
}

func NewImageModel(client *Client) *ImageModel {
	return &ImageModel{client: client}
}

type imageGenerationRequest struct {
	Model            string `json:"model"`
	Prompt           string `json:"prompt"`
	Quality          string `json:"quality,omitempty"`
	Size             string `json:"size,omitempty"`
	WatermarkEnabled *bool  `json:"watermark_enabled,omitempty"`
	UserID           string `json:"user_id,omitempty"`
}

func (m *ImageModel) CreateImageTask(ctx context.Context, prompt string, options core.AsyncImageOptions) (core.AsyncTask, error) {
	if prompt == "" {
		return core.AsyncTask{}, fmt.Errorf("prompt不能为空")
	}

	model := options.Model
	if model == "" {
		model = "glm-image"
	}

	req := imageGenerationRequest{
		Model:            model,
		Prompt:           prompt,
		Quality:          options.Quality,
		Size:             options.Size,
		WatermarkEnabled: options.WatermarkEnabled,
		UserID:           options.UserID,
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/async/images/generations", req)
	if err != nil {
		return core.AsyncTask{}, fmt.Errorf("创建图像生成任务失败: %w", err)
	}

	var resp asyncTaskResponse
	if err := jsonUnmarshal(respBody, &resp); err != nil {
		return core.AsyncTask{}, fmt.Errorf("反序列化响应失败: %w", err)
	}

	return toCoreTask(resp, "zhipu"), nil
}
