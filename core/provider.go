package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AsyncMediaProvider interface {
	Name() string
	ImageModel() AsyncImageModel
	VideoModel() AsyncVideoModel
	TaskQuery() AsyncTaskQuery
}

type ProviderConfig struct {
	Name    string
	APIKey  string
	BaseURL string
	Debug   bool
	Extra   map[string]interface{}
}

type ProviderFactory func(config ProviderConfig) (AsyncMediaProvider, error)

type ProviderRegistry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

func (r *ProviderRegistry) Register(name string, factory ProviderFactory) error {
	if name == "" {
		return fmt.Errorf("provider名称不能为空")
	}
	if factory == nil {
		return fmt.Errorf("provider工厂不能为空")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider '%s' 已注册", name)
	}

	r.factories[name] = factory
	return nil
}

func (r *ProviderRegistry) CreateProvider(name string, config ProviderConfig) (AsyncMediaProvider, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' 未注册", name)
	}

	config.Name = name
	return factory(config)
}

func (r *ProviderRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

func (r *ProviderRegistry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[name]
	return exists
}

func (r *ProviderRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; !exists {
		return fmt.Errorf("provider '%s' 未注册", name)
	}

	delete(r.factories, name)
	return nil
}

var globalRegistry = NewProviderRegistry()

func RegisterProvider(name string, factory ProviderFactory) error {
	return globalRegistry.Register(name, factory)
}

func CreateProvider(name string, config ProviderConfig) (AsyncMediaProvider, error) {
	return globalRegistry.CreateProvider(name, config)
}

func ListProviders() []string {
	return globalRegistry.ListProviders()
}

type MediaGenerationService struct {
	provider AsyncMediaProvider
}

func NewMediaGenerationService(provider AsyncMediaProvider) *MediaGenerationService {
	return &MediaGenerationService{provider: provider}
}

func (s *MediaGenerationService) Provider() AsyncMediaProvider {
	return s.provider
}

func (s *MediaGenerationService) SubmitImageTask(ctx context.Context, prompt string, options AsyncImageOptions) (AsyncTask, error) {
	return s.provider.ImageModel().CreateImageTask(ctx, prompt, options)
}

func (s *MediaGenerationService) SubmitVideoTask(ctx context.Context, prompt string, options AsyncVideoOptions) (AsyncTask, error) {
	return s.provider.VideoModel().CreateVideoTask(ctx, prompt, options)
}

func (s *MediaGenerationService) GenerateImage(ctx context.Context, prompt string, options AsyncImageOptions) (AsyncTaskResult, error) {
	task, err := s.provider.ImageModel().CreateImageTask(ctx, prompt, options)
	if err != nil {
		return AsyncTaskResult{}, fmt.Errorf("提交图像生成任务失败: %w", err)
	}

	return s.provider.TaskQuery().WaitForTask(ctx, task.ID, PollOptions{Interval: 5, MaxRetries: 60})
}

func (s *MediaGenerationService) GenerateVideo(ctx context.Context, prompt string, options AsyncVideoOptions) (AsyncTaskResult, error) {
	task, err := s.provider.VideoModel().CreateVideoTask(ctx, prompt, options)
	if err != nil {
		return AsyncTaskResult{}, fmt.Errorf("提交视频生成任务失败: %w", err)
	}

	return s.provider.TaskQuery().WaitForTask(ctx, task.ID, PollOptions{Interval: 10, MaxRetries: 60})
}

func (s *MediaGenerationService) GenerateImageWithPoll(ctx context.Context, prompt string, options AsyncImageOptions, pollOptions PollOptions) (AsyncTaskResult, error) {
	task, err := s.provider.ImageModel().CreateImageTask(ctx, prompt, options)
	if err != nil {
		return AsyncTaskResult{}, fmt.Errorf("提交图像生成任务失败: %w", err)
	}

	return s.provider.TaskQuery().WaitForTask(ctx, task.ID, pollOptions)
}

func (s *MediaGenerationService) GenerateVideoWithPoll(ctx context.Context, prompt string, options AsyncVideoOptions, pollOptions PollOptions) (AsyncTaskResult, error) {
	task, err := s.provider.VideoModel().CreateVideoTask(ctx, prompt, options)
	if err != nil {
		return AsyncTaskResult{}, fmt.Errorf("提交视频生成任务失败: %w", err)
	}

	return s.provider.TaskQuery().WaitForTask(ctx, task.ID, pollOptions)
}

func (s *MediaGenerationService) QueryTask(ctx context.Context, taskID string) (AsyncTaskResult, error) {
	return s.provider.TaskQuery().QueryTask(ctx, taskID)
}

func (s *MediaGenerationService) WaitForTask(ctx context.Context, taskID string, options PollOptions) (AsyncTaskResult, error) {
	return s.provider.TaskQuery().WaitForTask(ctx, taskID, options)
}

func NewDefaultPollOptions() PollOptions {
	return PollOptions{
		Interval:   5,
		MaxRetries: 60,
	}
}

func NewImagePollOptions() PollOptions {
	return PollOptions{
		Interval:   5,
		MaxRetries: 40,
	}
}

func NewVideoPollOptions() PollOptions {
	return PollOptions{
		Interval:   10,
		MaxRetries: 60,
	}
}

func EstimatePollDuration(options PollOptions) time.Duration {
	return time.Duration(options.Interval*options.MaxRetries) * time.Second
}
