package zhipu

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
	defaultBaseURL = "https://open.bigmodel.cn/api/paas/v4"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	debug      bool
}

type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

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
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.ErrDetail.Message != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("API请求失败，状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type APIError struct {
	ErrDetail APIErrorDetail `json:"error"`
}

type APIErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("智谱API错误 [%s]: %s", e.ErrDetail.Code, e.ErrDetail.Message)
}

type asyncTaskResponse struct {
	Model      string `json:"model"`
	ID         string `json:"id"`
	RequestID  string `json:"request_id"`
	TaskStatus string `json:"task_status"`
}

type asyncResultResponse struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"request_id"`
	Created     int64                  `json:"created"`
	Model       string                 `json:"model"`
	TaskStatus  string                 `json:"task_status"`
	VideoResult []videoResultItem      `json:"video_result,omitempty"`
	Choices     []interface{}          `json:"choices,omitempty"`
	Usage       map[string]interface{} `json:"usage,omitempty"`
}

type videoResultItem struct {
	URL           string `json:"url"`
	CoverImageURL string `json:"cover_image_url"`
}

func toCoreTask(resp asyncTaskResponse, provider string) core.AsyncTask {
	return core.AsyncTask{
		ID:         resp.ID,
		RequestID:  resp.RequestID,
		Model:      resp.Model,
		TaskStatus: core.TaskStatus(resp.TaskStatus),
		Provider:   provider,
		Raw:        nil,
	}
}

func toCoreTaskResult(resp asyncResultResponse, provider string) core.AsyncTaskResult {
	task := core.AsyncTask{
		ID:         resp.ID,
		RequestID:  resp.RequestID,
		Model:      resp.Model,
		TaskStatus: core.TaskStatus(resp.TaskStatus),
		Provider:   provider,
		CreatedAt:  resp.Created,
	}

	result := core.AsyncTaskResult{
		Task: task,
	}

	for _, vr := range resp.VideoResult {
		result.VideoResults = append(result.VideoResults, core.VideoResult{
			URL:           vr.URL,
			CoverImageURL: vr.CoverImageURL,
		})
	}

	return result
}
