package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	// Create chat model
	chatModel := openai.NewChatModel(client)

	// Create chat client
	chatClient := core.NewChatClient(chatModel)

	// Example 1: Basic Streaming
	fmt.Println("=== Example 1: Basic Streaming ===")
	stream, err := chatClient.Prompt().
		User("Tell me a short story about a robot learning to paint.").
		StreamWithContext(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	defer stream.Close()

	fmt.Print("Response: ")
	for {
		chunk, err := stream.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Stream error: %v\n", err)
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println()

	// Example 2: Streaming with System Message
	fmt.Println("=== Example 2: Streaming with System Message ===")
	stream2, err := chatClient.Prompt().
		System("You are a creative writing assistant.").
		User("Write a haiku about programming.").
		StreamWithContext(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	defer stream2.Close()

	fmt.Print("Response: ")
	for {
		chunk, err := stream2.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Stream error: %v\n", err)
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println()

	// Example 3: Streaming with Options
	fmt.Println("=== Example 3: Streaming with Options ===")
	stream3, err := chatClient.Prompt().
		User("Explain quantum computing in simple terms.").
		WithModel("gpt-4").
		WithTemperature(0.7).
		WithMaxTokens(200).
		StreamWithContext(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	defer stream3.Close()

	fmt.Print("Response: ")
	for {
		chunk, err := stream3.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Stream error: %v\n", err)
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println()

	// Example 4: Streaming with Template Variables
	fmt.Println("=== Example 4: Streaming with Template Variables ===")
	stream4, err := chatClient.Prompt().
		User("Write a brief introduction about {{.Topic}} for beginners.").
		WithVariable("Topic", "machine learning").
		StreamWithContext(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	defer stream4.Close()

	fmt.Print("Response: ")
	for {
		chunk, err := stream4.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Stream error: %v\n", err)
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println()

	fmt.Println("Streaming examples completed!")
}
