package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("需要设置 OPENAI_API_KEY 环境变量")
	}

	ctx := context.Background()

	client := openai.NewClient(apiKey)
	chatModel := openai.NewChatModel(client)

	fmt.Println("=== 示例 1: 基本对话记忆 ===")
	memory := core.NewConversationMemory()

	_ = memory.Add(core.Message{Role: core.RoleUser, Content: "我叫小明。"})
	_ = memory.Add(core.Message{Role: core.RoleAssistant, Content: "你好，小明！"})

	messages, err := memory.Get()
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("对话历史 (%d 条消息):\n", len(messages))
		for _, msg := range messages {
			fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
		}
	}

	fmt.Println("\n=== 示例 2: 窗口化对话记忆 ===")
	windowMemory := core.NewWindowedConversationMemory(4)

	_ = windowMemory.Add(core.Message{Role: core.RoleUser, Content: "消息1"})
	_ = windowMemory.Add(core.Message{Role: core.RoleAssistant, Content: "回复1"})
	_ = windowMemory.Add(core.Message{Role: core.RoleUser, Content: "消息2"})
	_ = windowMemory.Add(core.Message{Role: core.RoleAssistant, Content: "回复2"})
	_ = windowMemory.Add(core.Message{Role: core.RoleUser, Content: "消息3"})
	_ = windowMemory.Add(core.Message{Role: core.RoleAssistant, Content: "回复3"})

	windowMessages, _ := windowMemory.Get()
	fmt.Printf("窗口记忆 (最多4条，当前 %d 条):\n", len(windowMessages))
	for _, msg := range windowMessages {
		fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
	}

	fmt.Println("\n=== 示例 3: 摘要对话记忆 ===")
	summaryMemory := core.NewSummaryConversationMemory(chatModel, 6)

	_ = summaryMemory.Add(core.Message{Role: core.RoleUser, Content: "我喜欢Go编程。"})
	_ = summaryMemory.Add(core.Message{Role: core.RoleAssistant, Content: "Go是一门很棒的语言！"})

	summaryMessages, _ := summaryMemory.Get()
	fmt.Printf("摘要记忆 (%d 条消息):\n", len(summaryMessages))
	for _, msg := range summaryMessages {
		fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
	}

	fmt.Println("\n=== 示例 4: 日志顾问 ===")
	loggerAdvisor := &core.SimpleLoggerAdvisor{
		Logger: func(format string, args ...interface{}) {
			fmt.Printf("[日志] "+format+"\n", args...)
		},
	}

	chatClient := core.NewChatClient(chatModel).
		Builder().
		WithDefaultAdvisors(loggerAdvisor).
		Build()

	response, err := chatClient.Prompt().
		User("你好！").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("回答: %s\n", response.Content())
	}

	fmt.Println("\n=== 示例 5: 清除记忆 ===")
	clearMemory := core.NewConversationMemory()
	_ = clearMemory.Add(core.Message{Role: core.RoleUser, Content: "记住这个。"})

	beforeClear, _ := clearMemory.Get()
	fmt.Printf("清除前: %d 条消息\n", len(beforeClear))

	_ = clearMemory.Clear()

	afterClear, _ := clearMemory.Get()
	fmt.Printf("清除后: %d 条消息\n", len(afterClear))
}
