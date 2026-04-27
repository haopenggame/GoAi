package zhipu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func newTestClient(handler http.Handler) *Client {
	server := httptest.NewServer(handler)
	return NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
}

func TestClient(t *testing.T) {
	t.Run("创建客户端", func(t *testing.T) {
		client := NewClient("test-key")
		assert.NotNil(t, client)
		assert.Equal(t, "test-key", client.apiKey)
		assert.Equal(t, defaultBaseURL, client.baseURL)
	})

	t.Run("自定义选项", func(t *testing.T) {
		client := NewClient("test-key",
			WithBaseURL("https://custom.api.com/v4/"),
			WithDebug(true),
		)
		assert.Equal(t, "https://custom.api.com/v4", client.baseURL)
		assert.True(t, client.debug)
	})
}

func TestImageModel(t *testing.T) {
	t.Run("创建图像模型", func(t *testing.T) {
		client := NewClient("test-key")
		model := NewImageModel(client)
		assert.NotNil(t, model)
	})

	t.Run("创建图像生成任务-成功", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.URL.Path, "/async/images/generations")

			var req imageGenerationRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "glm-image", req.Model)
			assert.Equal(t, "一只可爱的小猫咪", req.Prompt)

			resp := asyncTaskResponse{
				Model:      "glm-image",
				ID:         "img-task-001",
				RequestID:  "req-001",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		model := NewImageModel(client)

		task, err := model.CreateImageTask(context.Background(), "一只可爱的小猫咪", core.AsyncImageOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "img-task-001", task.ID)
		assert.Equal(t, "req-001", task.RequestID)
		assert.Equal(t, core.TaskProcessing, task.TaskStatus)
		assert.Equal(t, "zhipu", task.Provider)
	})

	t.Run("创建图像生成任务-自定义参数", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req imageGenerationRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "glm-image", req.Model)
			assert.Equal(t, "hd", req.Quality)
			assert.Equal(t, "1024x1024", req.Size)

			resp := asyncTaskResponse{
				Model:      "glm-image",
				ID:         "img-task-002",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		model := NewImageModel(client)

		task, err := model.CreateImageTask(context.Background(), "测试图片", core.AsyncImageOptions{
			Quality: "hd",
			Size:    "1024x1024",
		})
		assert.NoError(t, err)
		assert.Equal(t, "img-task-002", task.ID)
	})

	t.Run("创建图像生成任务-空prompt", func(t *testing.T) {
		client := NewClient("test-key")
		model := NewImageModel(client)

		_, err := model.CreateImageTask(context.Background(), "", core.AsyncImageOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prompt不能为空")
	})

	t.Run("创建图像生成任务-API错误", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			resp := APIError{
				ErrDetail: APIErrorDetail{
					Code:    "invalid_request",
					Message: "请求参数错误",
				},
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		model := NewImageModel(client)

		_, err := model.CreateImageTask(context.Background(), "测试", core.AsyncImageOptions{})
		assert.Error(t, err)
	})
}

func TestVideoModel(t *testing.T) {
	t.Run("创建视频模型", func(t *testing.T) {
		client := NewClient("test-key")
		model := NewVideoModel(client)
		assert.NotNil(t, model)
	})

	t.Run("创建视频生成任务-成功", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/videos/generations")

			var req videoGenerationRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "cogvideox-3", req.Model)
			assert.Equal(t, "A cat is playing", req.Prompt)

			resp := asyncTaskResponse{
				Model:      "cogvideox-3",
				ID:         "vid-task-001",
				RequestID:  "req-v001",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		model := NewVideoModel(client)

		task, err := model.CreateVideoTask(context.Background(), "A cat is playing", core.AsyncVideoOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "vid-task-001", task.ID)
		assert.Equal(t, core.TaskProcessing, task.TaskStatus)
		assert.Equal(t, "zhipu", task.Provider)
	})

	t.Run("创建视频生成任务-自定义参数", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req videoGenerationRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "cogvideox-3", req.Model)
			assert.Equal(t, "quality", req.Quality)
			assert.Equal(t, "1920x1080", req.Size)
			assert.Equal(t, 30, req.FPS)
			assert.Equal(t, 5, req.Duration)

			resp := asyncTaskResponse{
				Model:      "cogvideox-3",
				ID:         "vid-task-002",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		model := NewVideoModel(client)

		withAudio := true
		task, err := model.CreateVideoTask(context.Background(), "测试视频", core.AsyncVideoOptions{
			Quality:   "quality",
			WithAudio: &withAudio,
			Size:      "1920x1080",
			FPS:       30,
			Duration:  5,
		})
		assert.NoError(t, err)
		assert.Equal(t, "vid-task-002", task.ID)
	})

	t.Run("创建视频生成任务-prompt和imageURL都为空", func(t *testing.T) {
		client := NewClient("test-key")
		model := NewVideoModel(client)

		_, err := model.CreateVideoTask(context.Background(), "", core.AsyncVideoOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prompt和image_url不能同时为空")
	})
}

func TestTaskQuery(t *testing.T) {
	t.Run("创建任务查询器", func(t *testing.T) {
		client := NewClient("test-key")
		query := NewTaskQuery(client)
		assert.NotNil(t, query)
	})

	t.Run("查询任务-处理中", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/async-result/task-001")

			resp := asyncResultResponse{
				ID:         "task-001",
				RequestID:  "req-001",
				Model:      "cogvideox-3",
				TaskStatus: "PROCESSING",
				Created:    1234567890,
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		query := NewTaskQuery(client)

		result, err := query.QueryTask(context.Background(), "task-001")
		assert.NoError(t, err)
		assert.Equal(t, "task-001", result.Task.ID)
		assert.Equal(t, core.TaskProcessing, result.Task.TaskStatus)
	})

	t.Run("查询任务-成功并返回视频结果", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := asyncResultResponse{
				ID:         "task-002",
				RequestID:  "req-002",
				Model:      "cogvideox-3",
				TaskStatus: "SUCCESS",
				Created:    1234567890,
				VideoResult: []videoResultItem{
					{
						URL:           "https://example.com/video.mp4",
						CoverImageURL: "https://example.com/cover.jpg",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		query := NewTaskQuery(client)

		result, err := query.QueryTask(context.Background(), "task-002")
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
		assert.Len(t, result.VideoResults, 1)
		assert.Equal(t, "https://example.com/video.mp4", result.VideoResults[0].URL)
		assert.Equal(t, "https://example.com/cover.jpg", result.VideoResults[0].CoverImageURL)
	})

	t.Run("查询任务-空taskID", func(t *testing.T) {
		client := NewClient("test-key")
		query := NewTaskQuery(client)

		_, err := query.QueryTask(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "taskID不能为空")
	})

	t.Run("等待任务-直接成功", func(t *testing.T) {
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			resp := asyncResultResponse{
				ID:         "task-003",
				TaskStatus: "SUCCESS",
				VideoResult: []videoResultItem{
					{URL: "https://example.com/video.mp4"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		query := NewTaskQuery(client)

		result, err := query.WaitForTask(context.Background(), "task-003", core.PollOptions{Interval: 1, MaxRetries: 5})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
		assert.Equal(t, 1, callCount)
	})

	t.Run("等待任务-从处理中到成功", func(t *testing.T) {
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount < 3 {
				resp := asyncResultResponse{
					ID:         "task-004",
					TaskStatus: "PROCESSING",
				}
				json.NewEncoder(w).Encode(resp)
			} else {
				resp := asyncResultResponse{
					ID:         "task-004",
					TaskStatus: "SUCCESS",
					VideoResult: []videoResultItem{
						{URL: "https://example.com/video.mp4"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			}
		})

		client := newTestClient(handler)
		query := NewTaskQuery(client)

		result, err := query.WaitForTask(context.Background(), "task-004", core.PollOptions{Interval: 1, MaxRetries: 10})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
		assert.Equal(t, 3, callCount)
	})

	t.Run("等待任务-任务失败", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := asyncResultResponse{
				ID:         "task-005",
				TaskStatus: "FAIL",
			}
			json.NewEncoder(w).Encode(resp)
		})

		client := newTestClient(handler)
		query := NewTaskQuery(client)

		_, err := query.WaitForTask(context.Background(), "task-005", core.PollOptions{Interval: 1, MaxRetries: 5})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "异步任务失败")
	})
}

func TestProvider(t *testing.T) {
	t.Run("创建Provider", func(t *testing.T) {
		provider := NewProvider("test-key")
		assert.NotNil(t, provider)
		assert.Equal(t, ProviderName, provider.Name())
		assert.NotNil(t, provider.ImageModel())
		assert.NotNil(t, provider.VideoModel())
		assert.NotNil(t, provider.TaskQuery())
	})

	t.Run("Provider实现AsyncMediaProvider接口", func(t *testing.T) {
		var _ core.AsyncMediaProvider = NewProvider("test-key")
	})

	t.Run("Provider自定义选项", func(t *testing.T) {
		provider := NewProvider("test-key",
			WithBaseURL("https://custom.api.com/v4"),
			WithDebug(true),
		)
		assert.NotNil(t, provider)
		assert.Equal(t, "zhipu", provider.Name())
	})

	t.Run("Provider图像生成端到端", func(t *testing.T) {
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method == "POST" {
				resp := asyncTaskResponse{
					Model:      "glm-image",
					ID:         "e2e-img-001",
					TaskStatus: "PROCESSING",
				}
				json.NewEncoder(w).Encode(resp)
			} else {
				resp := asyncResultResponse{
					ID:         "e2e-img-001",
					TaskStatus: "SUCCESS",
				}
				json.NewEncoder(w).Encode(resp)
			}
		})

		server := httptest.NewServer(handler)
		provider := NewProvider("test-key",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
		)

		task, err := provider.ImageModel().CreateImageTask(context.Background(), "测试图片", core.AsyncImageOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "e2e-img-001", task.ID)

		result, err := provider.TaskQuery().QueryTask(context.Background(), task.ID)
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})

	t.Run("Provider视频生成端到端", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				resp := asyncTaskResponse{
					Model:      "cogvideox-3",
					ID:         "e2e-vid-001",
					TaskStatus: "PROCESSING",
				}
				json.NewEncoder(w).Encode(resp)
			} else {
				resp := asyncResultResponse{
					ID:         "e2e-vid-001",
					TaskStatus: "SUCCESS",
					VideoResult: []videoResultItem{
						{URL: "https://example.com/video.mp4", CoverImageURL: "https://example.com/cover.jpg"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			}
		})

		server := httptest.NewServer(handler)
		provider := NewProvider("test-key",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
		)

		task, err := provider.VideoModel().CreateVideoTask(context.Background(), "测试视频", core.AsyncVideoOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "e2e-vid-001", task.ID)

		result, err := provider.TaskQuery().WaitForTask(context.Background(), task.ID, core.PollOptions{Interval: 1, MaxRetries: 5})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
		assert.Len(t, result.VideoResults, 1)
	})
}

func TestMediaGenerationService(t *testing.T) {
	t.Run("通过Provider创建Service", func(t *testing.T) {
		provider := NewProvider("test-key")
		service := core.NewMediaGenerationService(provider)
		assert.NotNil(t, service)
		assert.Equal(t, provider, service.Provider())
	})

	t.Run("Service提交图像任务", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := asyncTaskResponse{
				Model:      "glm-image",
				ID:         "svc-img-001",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		server := httptest.NewServer(handler)
		provider := NewProvider("test-key",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
		)
		service := core.NewMediaGenerationService(provider)

		task, err := service.SubmitImageTask(context.Background(), "测试图片", core.AsyncImageOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "svc-img-001", task.ID)
	})

	t.Run("Service提交视频任务", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := asyncTaskResponse{
				Model:      "cogvideox-3",
				ID:         "svc-vid-001",
				TaskStatus: "PROCESSING",
			}
			json.NewEncoder(w).Encode(resp)
		})

		server := httptest.NewServer(handler)
		provider := NewProvider("test-key",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
		)
		service := core.NewMediaGenerationService(provider)

		task, err := service.SubmitVideoTask(context.Background(), "测试视频", core.AsyncVideoOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "svc-vid-001", task.ID)
	})

	t.Run("Service生成图像(提交+等待)", func(t *testing.T) {
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method == "POST" {
				resp := asyncTaskResponse{
					Model:      "glm-image",
					ID:         "svc-img-002",
					TaskStatus: "PROCESSING",
				}
				json.NewEncoder(w).Encode(resp)
			} else {
				resp := asyncResultResponse{
					ID:         "svc-img-002",
					TaskStatus: "SUCCESS",
				}
				json.NewEncoder(w).Encode(resp)
			}
		})

		server := httptest.NewServer(handler)
		provider := NewProvider("test-key",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
		)
		service := core.NewMediaGenerationService(provider)

		result, err := service.GenerateImageWithPoll(context.Background(), "测试图片", core.AsyncImageOptions{}, core.PollOptions{Interval: 1, MaxRetries: 5})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})
}

func TestAPIError(t *testing.T) {
	t.Run("API错误格式化", func(t *testing.T) {
		apiErr := &APIError{
			ErrDetail: APIErrorDetail{
				Code:    "rate_limit",
				Message: "请求频率超限",
			},
		}
		assert.Contains(t, apiErr.Error(), "rate_limit")
		assert.Contains(t, apiErr.Error(), "请求频率超限")
	})
}

func TestCoreTypes(t *testing.T) {
	t.Run("任务状态常量", func(t *testing.T) {
		assert.Equal(t, core.TaskStatus("PROCESSING"), core.TaskProcessing)
		assert.Equal(t, core.TaskStatus("SUCCESS"), core.TaskSuccess)
		assert.Equal(t, core.TaskStatus("FAIL"), core.TaskFail)
	})

	t.Run("异步图像选项", func(t *testing.T) {
		watermark := false
		opts := core.AsyncImageOptions{
			Model:            "glm-image",
			Size:             "1280x1280",
			Quality:          "hd",
			WatermarkEnabled: &watermark,
			UserID:           "user-001",
		}
		assert.Equal(t, "glm-image", opts.Model)
		assert.Equal(t, "1280x1280", opts.Size)
		assert.Equal(t, "hd", opts.Quality)
		assert.False(t, *opts.WatermarkEnabled)
	})

	t.Run("异步视频选项", func(t *testing.T) {
		withAudio := true
		opts := core.AsyncVideoOptions{
			Model:     "cogvideox-3",
			Quality:   "quality",
			WithAudio: &withAudio,
			Size:      "1920x1080",
			FPS:       30,
			Duration:  5,
			UserID:    "user-001",
		}
		assert.Equal(t, "cogvideox-3", opts.Model)
		assert.True(t, *opts.WithAudio)
		assert.Equal(t, 30, opts.FPS)
	})

	t.Run("异步任务", func(t *testing.T) {
		task := core.AsyncTask{
			ID:         "task-001",
			RequestID:  "req-001",
			Model:      "glm-image",
			TaskStatus: core.TaskProcessing,
			Provider:   "zhipu",
		}
		assert.Equal(t, "task-001", task.ID)
		assert.Equal(t, "zhipu", task.Provider)
	})

	t.Run("轮询选项", func(t *testing.T) {
		opts := core.PollOptions{
			Interval:   3,
			MaxRetries: 20,
		}
		assert.Equal(t, 3, opts.Interval)
		assert.Equal(t, 20, opts.MaxRetries)
	})

	t.Run("Provider配置", func(t *testing.T) {
		config := core.ProviderConfig{
			Name:    "zhipu",
			APIKey:  "test-key",
			BaseURL: "https://custom.api.com",
			Debug:   true,
			Extra:   map[string]interface{}{"timeout": 30},
		}
		assert.Equal(t, "zhipu", config.Name)
		assert.Equal(t, "test-key", config.APIKey)
		assert.Equal(t, "https://custom.api.com", config.BaseURL)
		assert.True(t, config.Debug)
		assert.Equal(t, 30, config.Extra["timeout"])
	})
}
