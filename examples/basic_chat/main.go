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
	apiKey = "sk-xhCCBoqtOGU58EnylendD5ls9IOm75OWecp9LV5dBleqR5fQ"
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey, openai.WithBaseURL("https://api.hunyuan.cloud.tencent.com"))

	// Create chat model
	chatModel := openai.NewChatModel(client)

	// Create chat client with fluent API
	chatClient := core.NewChatClient(chatModel)
    
	ctx := context.Background()

	// Example 1: Simple chat
	fmt.Println("=== Example 1: Simple Chat ===")
	response, err := chatClient.Prompt().WithModel("hunyuan-lite").
		User("What is the capital of France?").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response.Content())
	}

	// Example 2: Chat with system message
	fmt.Println("=== Example 2: Chat with System Message ===")
	response, err = chatClient.Prompt().
		System("You are a helpful coding assistant.").
		User("Write a simple Hello World program in Go.").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response.Content())
	}

	// Example 3: Chat with options
	fmt.Println("=== Example 3: Chat with Options ===")
	response, err = chatClient.Prompt().
		User("Tell me a joke.").
		WithModel("gpt-4").
		WithTemperature(0.8).
		WithMaxTokens(100).
		CallWithContext(ctx)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response.Content())
	}

	// Example 4: Chat with template variables
	fmt.Println("=== Example 4: Chat with Template Variables ===")
	response, err = chatClient.Prompt().
		User("Tell me about {{.Topic}} in {{.Language}}.").
		WithVariable("Topic", "concurrency").
		WithVariable("Language", "Go").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response.Content())
	}
}
