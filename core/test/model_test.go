package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestChatOptions(t *testing.T) {
	t.Run("创建聊天选项", func(t *testing.T) {
		options := core.ChatOptions{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   1000,
		}
		assert.Equal(t, "gpt-4", options.Model)
		assert.Equal(t, 0.7, options.Temperature)
		assert.Equal(t, 1000, options.MaxTokens)
	})
}

func TestMessage(t *testing.T) {
	t.Run("创建消息", func(t *testing.T) {
		message := core.Message{
			Role:    core.RoleUser,
			Content: "Hello, world!",
		}
		assert.Equal(t, core.RoleUser, message.Role)
		assert.Equal(t, "Hello, world!", message.Content)
	})
}

func TestChatResponse(t *testing.T) {
	t.Run("创建聊天响应", func(t *testing.T) {
		response := core.ChatResponse{
			Result: core.GenerationResult{
				Output: core.Message{
					Role:    core.RoleAssistant,
					Content: "Hello, world!",
				},
			},
		}
		assert.NotNil(t, response)
		assert.Equal(t, core.RoleAssistant, response.Result.Output.Role)
		assert.Equal(t, "Hello, world!", response.Result.Output.Content)
	})
}
