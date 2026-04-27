package zhipu

import (
	"context"
	"fmt"

	"github.com/go-spring/ai/core"
)

type VideoModel struct {
	client *Client
}

func NewVideoModel(client *Client) *VideoModel {
	return &VideoModel{client: client}
}

type videoGenerationRequest struct {
	Model            string      `json:"model"`
	Prompt           string      `json:"prompt,omitempty"`
	Quality          string      `json:"quality,omitempty"`
	WithAudio        *bool       `json:"with_audio,omitempty"`
	WatermarkEnabled *bool       `json:"watermark_enabled,omitempty"`
	ImageURL         interface{} `json:"image_url,omitempty"`
	Size             string      `json:"size,omitempty"`
	FPS              int         `json:"fps,omitempty"`
	Duration         int         `json:"duration,omitempty"`
	RequestID        string      `json:"request_id,omitempty"`
	UserID           string      `json:"user_id,omitempty"`
}

func (m *VideoModel) CreateVideoTask(ctx context.Context, prompt string, options core.AsyncVideoOptions) (core.AsyncTask, error) {
	if prompt == "" && options.ImageURL == nil {
		return core.AsyncTask{}, fmt.Errorf("prompt和image_url不能同时为空")
	}

	model := options.Model
	if model == "" {
		model = "cogvideox-3"
	}

	req := videoGenerationRequest{
		Model:            model,
		Prompt:           prompt,
		Quality:          options.Quality,
		WithAudio:        options.WithAudio,
		WatermarkEnabled: options.WatermarkEnabled,
		ImageURL:         options.ImageURL,
		Size:             options.Size,
		FPS:              options.FPS,
		Duration:         options.Duration,
		RequestID:        options.RequestID,
		UserID:           options.UserID,
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/videos/generations", req)
	if err != nil {
		return core.AsyncTask{}, fmt.Errorf("创建视频生成任务失败: %w", err)
	}

	var resp asyncTaskResponse
	if err := jsonUnmarshal(respBody, &resp); err != nil {
		return core.AsyncTask{}, fmt.Errorf("反序列化响应失败: %w", err)
	}

	return toCoreTask(resp, "zhipu"), nil
}
