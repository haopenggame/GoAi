package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestChatClientResponse(t *testing.T) {
	// 创建一个简单的模拟模型
	model := &MockChatModel{response: "Hello, World!"}
	client := core.NewChatClient(model)

	t.Run("响应内容", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello").
			Call()
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("响应消息", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello").
			Call()
		assert.NoError(t, err)
		assert.NotNil(t, response.Message())
	})

	t.Run("响应元数据", func(t *testing.T) {
		response, err := client.Prompt().
			User("Hello").
			Call()
		assert.NoError(t, err)
		assert.NotNil(t, response.Metadata())
	})
}

func TestToolCalls(t *testing.T) {
	// 工具调用测试需要一个能够返回工具调用的模拟模型
	// 这里我们暂时跳过，因为需要更复杂的模拟实现
	t.Skip("工具调用测试需要更复杂的模拟实现")
}
