package zhipu

import (
	"fmt"

	"github.com/go-spring/ai/core"
)

const ProviderName = "zhipu"

type Provider struct {
	client     *Client
	imageModel *ImageModel
	videoModel *VideoModel
	taskQuery  *TaskQuery
}

func NewProvider(apiKey string, options ...ClientOption) *Provider {
	client := NewClient(apiKey, options...)
	return &Provider{
		client:     client,
		imageModel: NewImageModel(client),
		videoModel: NewVideoModel(client),
		taskQuery:  NewTaskQuery(client),
	}
}

func (p *Provider) Name() string {
	return ProviderName
}

func (p *Provider) ImageModel() core.AsyncImageModel {
	return p.imageModel
}

func (p *Provider) VideoModel() core.AsyncVideoModel {
	return p.videoModel
}

func (p *Provider) TaskQuery() core.AsyncTaskQuery {
	return p.taskQuery
}

func init() {
	_ = core.RegisterProvider(ProviderName, func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
		if config.APIKey == "" {
			return nil, fmt.Errorf("智谱API密钥不能为空")
		}

		var options []ClientOption
		if config.BaseURL != "" {
			options = append(options, WithBaseURL(config.BaseURL))
		}
		if config.Debug {
			options = append(options, WithDebug(true))
		}

		return NewProvider(config.APIKey, options...), nil
	})
}
