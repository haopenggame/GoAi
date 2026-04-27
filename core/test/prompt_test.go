package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestPromptTemplate(t *testing.T) {
	t.Run("创建简单模板", func(t *testing.T) {
		template := core.NewPromptTemplate("Hello, {{name}}!")
		result, err := template.Render(map[string]any{
			"name": "World",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("创建带多个变量的模板", func(t *testing.T) {
		template := core.NewPromptTemplate("{{greeting}}, {{name}}! You are {{age}} years old.")
		result, err := template.Render(map[string]any{
			"greeting": "Hi",
			"name":     "Alice",
			"age":      30,
		})
		assert.NoError(t, err)
		assert.Equal(t, "Hi, Alice! You are 30 years old.", result)
	})

	t.Run("创建空模板", func(t *testing.T) {
		template := core.NewPromptTemplate("")
		result, err := template.Render(map[string]any{})
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestChatClientPromptBuilder(t *testing.T) {
	t.Run("构建聊天提示词", func(t *testing.T) {
		// 创建一个简单的模拟模型
		model := &MockChatModel{response: "Hello, World!"}
		client := core.NewChatClient(model)

		// 构建提示词
		builder := client.Prompt()
		builder.System("You are a helpful assistant")
		builder.User("Hello")

		// 执行调用
		response, err := builder.Call()
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("带变量的提示词", func(t *testing.T) {
		model := &MockChatModel{response: "Hello, World!"}
		client := core.NewChatClient(model)

		response, err := client.Prompt().
			User("Hello {{.Name}}").
			WithVariable("Name", "Alice").
			Call()
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})

	t.Run("带选项的提示词", func(t *testing.T) {
		model := &MockChatModel{response: "Hello, World!"}
		client := core.NewChatClient(model)

		response, err := client.Prompt().
			User("Hello").
			WithModel("gpt-4").
			WithTemperature(0.7).
			WithMaxTokens(100).
			Call()
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", response.Content())
	})
}
