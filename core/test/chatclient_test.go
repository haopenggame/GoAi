package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

type MockChatModel struct {
	response string
}

func (m *MockChatModel) Call(ctx context.Context, prompt core.Prompt) (core.ChatResponse, error) {
	return core.ChatResponse{
		Result: core.GenerationResult{
			Output: core.Message{Content: m.response},
		},
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, prompt core.Prompt) (core.ChatStream, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestChatClient(t *testing.T) {
	ctx := context.Background()
	model := &MockChatModel{response: "Hello, World!"}
	client := core.NewChatClient(model)

	t.Run("基本聊天", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello").
			CallWithContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("带系统消息", func(t *testing.T) {
		response, err := client.Prompt().
			System("You are a helpful assistant").
			User("Hello").
			CallWithContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("带变量", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello {{.Name}}").
			WithVariable("Name", "Alice").
			CallWithContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("带选项", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello").
			WithModel("gpt-4").
			WithTemperature(0.7).
			WithMaxTokens(100).
			CallWithContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("流式响应", func(t *testing.T) {
		stream, err := client.Prompt().
			User("Hello").
			StreamWithContext(ctx)
		assert.Error(t, err) // Mock 模型未实现流式
		assert.Nil(t, stream)
	})

	t.Run("使用构建器", func(t *testing.T) {
		builder := client.Builder()
		builder.WithDefaultSystem("Default system message")
		newClient := builder.Build()

		response, err := newClient.Prompt().
			User("Hello").
			CallWithContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})
}

func TestSimpleLoggerAdvisor(t *testing.T) {
	ctx := context.Background()
	var loggedMessages []string

	advisor := &core.SimpleLoggerAdvisor{
		Logger: func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		},
	}

	prompt := core.Prompt{
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "Hello!"},
		},
	}
	advisorContext := make(map[string]interface{})

	_, err := advisor.Before(ctx, prompt, advisorContext)
	assert.NoError(t, err)
	assert.Len(t, loggedMessages, 1)
	assert.Contains(t, loggedMessages[0], "聊天请求")

	response := core.ChatResponse{
		Result: core.GenerationResult{
			Output: core.Message{Content: "Hi!"},
		},
	}

	modifiedResponse, err := advisor.After(ctx, response, advisorContext)
	assert.NoError(t, err)
	assert.Len(t, loggedMessages, 2)
	assert.Contains(t, loggedMessages[1], "聊天响应")
	assert.Equal(t, "Hi!", modifiedResponse.Result.Output.Content)
}
