package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/go-spring/ai/core"
)

// AudioModel 实现 core.AudioModel 的 OpenAI 音频模型
type AudioModel struct {
	client *Client
}

// NewAudioModel 创建新的 OpenAI 音频模型
func NewAudioModel(client *Client) *AudioModel {
	return &AudioModel{client: client}
}

// TranscriptionRequest 表示转录请求
type TranscriptionRequest struct {
	Model          string `json:"model"`
	Language       string `json:"language,omitempty"`
	Prompt         string `json:"prompt,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

// TranscriptionResponse 表示转录响应
type TranscriptionResponse struct {
	Text string `json:"text"`
}

// Transcribe 实现 core.AudioModel 接口
func (m *AudioModel) Transcribe(ctx context.Context, audio []byte, options core.AudioOptions) (string, error) {
	model := options.Model
	if model == "" {
		model = "whisper-1"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("创建表单文件失败: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(audio)); err != nil {
		return "", fmt.Errorf("复制音频数据失败: %w", err)
	}

	// 添加其他字段
	if err := writer.WriteField("model", model); err != nil {
		return "", err
	}
	if options.Language != "" {
		if err := writer.WriteField("language", options.Language); err != nil {
			return "", err
		}
	}
	if options.Prompt != "" {
		if err := writer.WriteField("prompt", options.Prompt); err != nil {
			return "", err
		}
	}
	if options.ResponseFormat != "" {
		if err := writer.WriteField("response_format", options.ResponseFormat); err != nil {
			return "", err
		}
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.client.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.client.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := m.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("转录请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("转录失败，状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp TranscriptionResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("反序列化响应失败: %w", err)
	}

	return apiResp.Text, nil
}

// SpeechRequest 表示语音合成请求
type SpeechRequest struct {
	Model string  `json:"model"`
	Input string  `json:"input"`
	Voice string  `json:"voice"`
	Speed float32 `json:"speed,omitempty"`
}

// Synthesize 实现 core.AudioModel 接口
func (m *AudioModel) Synthesize(ctx context.Context, text string, options core.AudioOptions) ([]byte, error) {
	model := options.Model
	if model == "" {
		model = "tts-1"
	}

	voice := options.Voice
	if voice == "" {
		voice = "alloy"
	}

	req := SpeechRequest{
		Model: model,
		Input: text,
		Voice: voice,
		Speed: options.Speed,
	}

	respBody, err := m.client.doRequest(ctx, "POST", "/audio/speech", req)
	if err != nil {
		return nil, fmt.Errorf("语音合成失败: %w", err)
	}

	return respBody, nil
}
