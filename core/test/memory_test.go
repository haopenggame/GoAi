package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestConversationMemory(t *testing.T) {
	memory := core.NewConversationMemory()

	t.Run("添加和获取消息", func(t *testing.T) {
		err := memory.Add(core.Message{Role: core.RoleUser, Content: "Hello"})
		assert.NoError(t, err)
		err = memory.Add(core.Message{Role: core.RoleAssistant, Content: "Hi!"})
		assert.NoError(t, err)

		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, core.RoleUser, messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
		assert.Equal(t, core.RoleAssistant, messages[1].Role)
		assert.Equal(t, "Hi!", messages[1].Content)
	})

	t.Run("清空记忆", func(t *testing.T) {
		err := memory.Add(core.Message{Role: core.RoleUser, Content: "Test"})
		assert.NoError(t, err)

		err = memory.Clear()
		assert.NoError(t, err)

		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestWindowedConversationMemory(t *testing.T) {
	memory := core.NewWindowedConversationMemory(3)

	t.Run("窗口大小限制", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			err := memory.Add(core.Message{Role: core.RoleUser, Content: string(rune('A' + i))})
			assert.NoError(t, err)
		}

		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Len(t, messages, 3)
		assert.Equal(t, "C", messages[0].Content)
		assert.Equal(t, "D", messages[1].Content)
		assert.Equal(t, "E", messages[2].Content)
	})

	t.Run("清空记忆", func(t *testing.T) {
		err := memory.Clear()
		assert.NoError(t, err)

		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})
}

func TestSummaryConversationMemoryWithoutModel(t *testing.T) {
	memory := core.NewSummaryConversationMemory(nil, 4)

	t.Run("无模型时截断消息", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			err := memory.Add(core.Message{Role: core.RoleUser, Content: string(rune('A' + i))})
			assert.NoError(t, err)
		}

		messages, err := memory.Get()
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(messages), 4)
	})
}