package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

// WeatherArgs represents the arguments for the weather tool.
type WeatherArgs struct {
	Location string `json:"location"`
	Units    string `json:"units,omitempty"`
}

// CalculatorArgs represents the arguments for the calculator tool.
type CalculatorArgs struct {
	Expression string `json:"expression"`
}

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

	// Create tool registry
	registry := core.NewToolRegistry()

	// Register weather tool
	weatherTool := core.NewFunctionToolBuilder("get_weather").
		WithDescription("Get the current weather for a location.").
		WithParameter("location", "string", "The city and state, e.g., San Francisco, CA", true).
		WithParameter("units", "string", "The temperature unit (celsius or fahrenheit)", false).
		Build()

	_ = registry.Register(weatherTool, func(ctx context.Context, arguments string) (string, error) {
		var args WeatherArgs
		if err := core.ParseArguments(arguments, &args); err != nil {
			return "", err
		}

		// In a real application, you would call a weather API here
		return fmt.Sprintf("The weather in %s is sunny and 25°C", args.Location), nil
	})

	// Register calculator tool
	calculatorTool := core.NewFunctionToolBuilder("calculate").
		WithDescription("Evaluate a mathematical expression.").
		WithParameter("expression", "string", "The mathematical expression to evaluate, e.g., '2 + 2'", true).
		Build()

	_ = registry.Register(calculatorTool, func(ctx context.Context, arguments string) (string, error) {
		var args CalculatorArgs
		if err := core.ParseArguments(arguments, &args); err != nil {
			return "", err
		}

		// In a real application, you would evaluate the expression safely
		return fmt.Sprintf("The result of '%s' is 4", args.Expression), nil
	})

	// Create tool-calling chat model
	toolModel := core.NewToolCallingChatModel(chatModel, registry)

	// Example 1: Simple tool usage
	fmt.Println("=== Example 1: Tool Usage ===")
	prompt := core.NewPromptBuilder().
		User("What's the weather like in Beijing?").
		Build()

	response, err := toolModel.Call(ctx, prompt)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response.Result.Output.Content)
	}

	// Example 2: Multiple tools
	fmt.Println("=== Example 2: Multiple Tools ===")
	prompt2 := core.NewPromptBuilder().
		User("What's the weather in Tokyo and what is 15 * 23?").
		Build()

	response2, err := toolModel.Call(ctx, prompt2)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", response2.Result.Output.Content)
	}

	// Example 3: Tool registry management
	fmt.Println("=== Example 3: Tool Registry ===")
	definitions := registry.GetAllDefinitions()
	fmt.Printf("Registered tools: %d\n", len(definitions))
	for _, def := range definitions {
		fmt.Printf("- %s: %s\n", def.Function.Name, def.Function.Description)
	}

	// Example 4: Manual tool execution
	fmt.Println("\n=== Example 4: Manual Tool Execution ===")
	toolCall := core.ToolCall{
		ID:   "call_manual",
		Type: "function",
		Function: core.FunctionCall{
			Name:      "get_weather",
			Arguments: `{"location": "Shanghai", "units": "celsius"}`,
		},
	}

	result, err := registry.Execute(ctx, toolCall)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Tool result: %s\n", result)
	}
}
