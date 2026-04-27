package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestSmartConversationMemory(t *testing.T) {
	t.Run("创建对话记忆", func(t *testing.T) {
		memory := core.NewConversationMemory()
		assert.NotNil(t, memory)
	})

	t.Run("添加和获取消息", func(t *testing.T) {
		memory := core.NewConversationMemory()

		// 添加消息
		message := core.Message{
			Role:    core.RoleUser,
			Content: "Hello",
		}
		err := memory.Add(message)
		assert.NoError(t, err)

		// 获取消息
		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, core.RoleUser, messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
	})

	t.Run("清除消息", func(t *testing.T) {
		memory := core.NewConversationMemory()

		// 添加消息
		message := core.Message{
			Role:    core.RoleUser,
			Content: "Hello",
		}
		err := memory.Add(message)
		assert.NoError(t, err)

		// 清除消息
		err = memory.Clear()
		assert.NoError(t, err)

		// 验证消息已清除
		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestSmartWindowedConversationMemory(t *testing.T) {
	t.Run("创建窗口化对话记忆", func(t *testing.T) {
		memory := core.NewWindowedConversationMemory(2)
		assert.NotNil(t, memory)
	})

	t.Run("窗口大小限制", func(t *testing.T) {
		memory := core.NewWindowedConversationMemory(2)

		// 添加3条消息
		messages := []core.Message{
			{Role: core.RoleUser, Content: "Message 1"},
			{Role: core.RoleAssistant, Content: "Response 1"},
			{Role: core.RoleUser, Content: "Message 2"},
		}

		for _, msg := range messages {
			err := memory.Add(msg)
			assert.NoError(t, err)
		}

		// 获取消息，应该只返回最后2条
		result, err := memory.Get()
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Response 1", result[0].Content)
		assert.Equal(t, "Message 2", result[1].Content)
	})
}

func TestSummaryConversationMemory(t *testing.T) {
	t.Run("创建摘要对话记忆", func(t *testing.T) {
		// 创建一个简单的模拟模型
		model := &MockChatModel{response: "摘要内容"}
		memory := core.NewSummaryConversationMemory(model, 3)
		assert.NotNil(t, memory)
	})

	t.Run("摘要功能", func(t *testing.T) {
		// 创建一个简单的模拟模型
		model := &MockChatModel{response: "摘要内容"}
		memory := core.NewSummaryConversationMemory(model, 3)

		// 添加4条消息，超过最大限制
		messages := []core.Message{
			{Role: core.RoleUser, Content: "Message 1"},
			{Role: core.RoleAssistant, Content: "Response 1"},
			{Role: core.RoleUser, Content: "Message 2"},
			{Role: core.RoleAssistant, Content: "Response 2"},
		}

		for _, msg := range messages {
			err := memory.Add(msg)
			assert.NoError(t, err)
		}

		// 获取消息，应该包含摘要
		result, err := memory.Get()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
	})
}
