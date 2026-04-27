package test

import (
	"context"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

type MockAsyncMediaProvider struct {
	name       string
	imageModel core.AsyncImageModel
	videoModel core.AsyncVideoModel
	taskQuery  core.AsyncTaskQuery
}

func (p *MockAsyncMediaProvider) Name() string {
	return p.name
}

func (p *MockAsyncMediaProvider) ImageModel() core.AsyncImageModel {
	return p.imageModel
}

func (p *MockAsyncMediaProvider) VideoModel() core.AsyncVideoModel {
	return p.videoModel
}

func (p *MockAsyncMediaProvider) TaskQuery() core.AsyncTaskQuery {
	return p.taskQuery
}

type MockAsyncImageModel struct{}

func (m *MockAsyncImageModel) CreateImageTask(ctx context.Context, prompt string, options core.AsyncImageOptions) (core.AsyncTask, error) {
	return core.AsyncTask{
		ID:         "mock-img-task",
		Model:      "mock-image",
		TaskStatus: core.TaskProcessing,
		Provider:   "mock",
	}, nil
}

type MockAsyncVideoModel struct{}

func (m *MockAsyncVideoModel) CreateVideoTask(ctx context.Context, prompt string, options core.AsyncVideoOptions) (core.AsyncTask, error) {
	return core.AsyncTask{
		ID:         "mock-vid-task",
		Model:      "mock-video",
		TaskStatus: core.TaskProcessing,
		Provider:   "mock",
	}, nil
}

type MockAsyncTaskQuery struct{}

func (q *MockAsyncTaskQuery) QueryTask(ctx context.Context, taskID string) (core.AsyncTaskResult, error) {
	return core.AsyncTaskResult{
		Task: core.AsyncTask{
			ID:         taskID,
			TaskStatus: core.TaskSuccess,
		},
	}, nil
}

func (q *MockAsyncTaskQuery) WaitForTask(ctx context.Context, taskID string, options core.PollOptions) (core.AsyncTaskResult, error) {
	return core.AsyncTaskResult{
		Task: core.AsyncTask{
			ID:         taskID,
			TaskStatus: core.TaskSuccess,
		},
	}, nil
}

func TestProviderRegistry(t *testing.T) {
	t.Run("创建注册中心", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		assert.NotNil(t, registry)
	})

	t.Run("注册Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		err := registry.Register("mock", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "mock"}, nil
		})
		assert.NoError(t, err)
		assert.True(t, registry.HasProvider("mock"))
	})

	t.Run("重复注册Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		_ = registry.Register("mock", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "mock"}, nil
		})
		err := registry.Register("mock", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "mock"}, nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已注册")
	})

	t.Run("注册空名称", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		err := registry.Register("", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return nil, nil
		})
		assert.Error(t, err)
	})

	t.Run("注册空工厂", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		err := registry.Register("mock", nil)
		assert.Error(t, err)
	})

	t.Run("创建Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		_ = registry.Register("mock", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: config.Name}, nil
		})

		provider, err := registry.CreateProvider("mock", core.ProviderConfig{APIKey: "test-key"})
		assert.NoError(t, err)
		assert.Equal(t, "mock", provider.Name())
	})

	t.Run("创建未注册的Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		_, err := registry.CreateProvider("not-exist", core.ProviderConfig{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "未注册")
	})

	t.Run("列出所有Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		_ = registry.Register("provider-a", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "a"}, nil
		})
		_ = registry.Register("provider-b", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "b"}, nil
		})

		names := registry.ListProviders()
		assert.Len(t, names, 2)
	})

	t.Run("注销Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		_ = registry.Register("mock", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
			return &MockAsyncMediaProvider{name: "mock"}, nil
		})

		err := registry.Unregister("mock")
		assert.NoError(t, err)
		assert.False(t, registry.HasProvider("mock"))
	})

	t.Run("注销未注册的Provider", func(t *testing.T) {
		registry := core.NewProviderRegistry()
		err := registry.Unregister("not-exist")
		assert.Error(t, err)
	})
}

func TestMediaGenerationService(t *testing.T) {
	provider := &MockAsyncMediaProvider{
		name:       "mock",
		imageModel: &MockAsyncImageModel{},
		videoModel: &MockAsyncVideoModel{},
		taskQuery:  &MockAsyncTaskQuery{},
	}

	t.Run("创建Service", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		assert.NotNil(t, service)
		assert.Equal(t, provider, service.Provider())
	})

	t.Run("提交图像任务", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		task, err := service.SubmitImageTask(context.Background(), "测试图片", core.AsyncImageOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "mock-img-task", task.ID)
		assert.Equal(t, core.TaskProcessing, task.TaskStatus)
	})

	t.Run("提交视频任务", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		task, err := service.SubmitVideoTask(context.Background(), "测试视频", core.AsyncVideoOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "mock-vid-task", task.ID)
		assert.Equal(t, core.TaskProcessing, task.TaskStatus)
	})

	t.Run("生成图像(提交+等待)", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		result, err := service.GenerateImage(context.Background(), "测试图片", core.AsyncImageOptions{})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})

	t.Run("生成视频(提交+等待)", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		result, err := service.GenerateVideo(context.Background(), "测试视频", core.AsyncVideoOptions{})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})

	t.Run("查询任务", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		result, err := service.QueryTask(context.Background(), "task-001")
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})

	t.Run("等待任务", func(t *testing.T) {
		service := core.NewMediaGenerationService(provider)
		result, err := service.WaitForTask(context.Background(), "task-001", core.PollOptions{Interval: 1, MaxRetries: 5})
		assert.NoError(t, err)
		assert.Equal(t, core.TaskSuccess, result.Task.TaskStatus)
	})
}

func TestPollOptionsHelpers(t *testing.T) {
	t.Run("默认轮询选项", func(t *testing.T) {
		opts := core.NewDefaultPollOptions()
		assert.Equal(t, 5, opts.Interval)
		assert.Equal(t, 60, opts.MaxRetries)
	})

	t.Run("图像轮询选项", func(t *testing.T) {
		opts := core.NewImagePollOptions()
		assert.Equal(t, 5, opts.Interval)
		assert.Equal(t, 40, opts.MaxRetries)
	})

	t.Run("视频轮询选项", func(t *testing.T) {
		opts := core.NewVideoPollOptions()
		assert.Equal(t, 10, opts.Interval)
		assert.Equal(t, 60, opts.MaxRetries)
	})

	t.Run("估算轮询时长", func(t *testing.T) {
		opts := core.PollOptions{Interval: 5, MaxRetries: 60}
		duration := core.EstimatePollDuration(opts)
		assert.Equal(t, 300, int(duration.Seconds()))
	})
}

func TestAsyncMediaProviderInterface(t *testing.T) {
	t.Run("MockProvider实现接口", func(t *testing.T) {
		var _ core.AsyncMediaProvider = &MockAsyncMediaProvider{}
	})

	t.Run("接口方法返回正确类型", func(t *testing.T) {
		provider := &MockAsyncMediaProvider{
			name:       "test",
			imageModel: &MockAsyncImageModel{},
			videoModel: &MockAsyncVideoModel{},
			taskQuery:  &MockAsyncTaskQuery{},
		}

		var p core.AsyncMediaProvider = provider
		assert.Equal(t, "test", p.Name())
		assert.NotNil(t, p.ImageModel())
		assert.NotNil(t, p.VideoModel())
		assert.NotNil(t, p.TaskQuery())
	})
}
