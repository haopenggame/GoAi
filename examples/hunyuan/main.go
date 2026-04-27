package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

type WeatherArgs struct {
	Location string `json:"location"`
	Units    string `json:"units,omitempty"`
}

type CalculatorArgs struct {
	Expression string `json:"expression"`
}

func main() {
	// 混元模型 API 密钥
	apiKey := "sk-xhCCBoqtOGU58EnylendD5ls9IOm75OWecp9LV5dBleqR5fQ"
	if apiKey == "" {
		log.Fatal("API 密钥不能为空")
	}

	ctx := context.Background()

	// 创建 OpenAI 客户端，使用混元模型的 base URL
	client := openai.NewClient(apiKey,
		openai.WithBaseURL("https://api.hunyuan.cloud.tencent.com/v1"),
		openai.WithDebug(true),           // 启用调试模式，打印 HTTP 日志
		openai.WithModel("hunyuan-lite"), // 使用混元模型
	)

	// 创建聊天模型
	chatModel := openai.NewChatModel(client)

	// 创建聊天客户端
	// chatClient := core.NewChatClient(chatModel)

	// 创建工具注册表
	registry := core.NewToolRegistry()

	// 注册天气工具
	weatherTool := core.NewFunctionToolBuilder("get_weather").
		WithDescription("获取指定位置的当前天气信息").
		WithParameter("location", "string", "城市名称，例如：北京、上海", true).
		WithParameter("units", "string", "温度单位（celsius 或 fahrenheit）", false).
		Build()

	_ = registry.Register(weatherTool, func(ctx context.Context, arguments string) (string, error) {
		var args WeatherArgs
		if err := core.ParseArguments(arguments, &args); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s 的天气是晴天，25°C", args.Location), nil
	})

	// 注册计算器工具
	calculatorTool := core.NewFunctionToolBuilder("calculate").
		WithDescription("计算数学表达式的结果").
		WithParameter("expression", "string", "要计算的数学表达式，例如：15 * 23", true).
		Build()

	_ = registry.Register(calculatorTool, func(ctx context.Context, arguments string) (string, error) {
		var args CalculatorArgs
		if err := core.ParseArguments(arguments, &args); err != nil {
			return "", err
		}
		return fmt.Sprintf("'%s' 的结果是 345", args.Expression), nil
	})

	// 创建支持工具调用的聊天模型
	toolModel := core.NewToolCallingChatModel(chatModel, registry)

	// // 示例：基本聊天
	// fmt.Println("=== 混元模型测试 ===")
	// response, err := chatClient.Prompt().
	// 	System("你是一个智能助手，会用简洁友好的语言回答问题。").
	// 	User("你好，介绍一下自己。").
	// 	CallWithContext(ctx)
	// if err != nil {
	// 	log.Fatalf("请求失败: %v", err)
	// }

	// fmt.Printf("响应: %s\n", response.Content())

	// // 示例：流式响应
	// fmt.Println("\n=== 流式响应测试 ===")
	// stream, err := chatClient.Prompt().
	// 	User("写一首关于春天的诗。").
	// 	StreamWithContext(ctx)
	// if err != nil {
	// 	log.Fatalf("创建流失败: %v", err)
	// }
	// defer stream.Close()

	// fmt.Print("响应: ")
	// for {
	// 	chunk, err := stream.Next()
	// 	if err != nil {
	// 		break
	// 	}
	// 	fmt.Print(chunk.Content)
	// }
	// fmt.Println()

	// 示例：工具调用
	fmt.Println("\n=== 工具调用测试 ===")

	// 创建支持工具调用的聊天客户端
	toolChatClient := core.NewChatClient(toolModel)

	toolResp, err := toolChatClient.Prompt().
		User("北京今天的天气怎么样？请帮我计算一下 15 * 23 等于多少？").
		CallWithContext(ctx)
	if err != nil {
		log.Printf("工具调用失败: %v\n", err)
	} else {
		fmt.Printf("响应: %s\n", toolResp.Content())
	}

	// 示例：查看已注册的工具
	fmt.Println("\n=== 已注册的工具 ===")
	definitions := registry.GetAllDefinitions()
	fmt.Printf("共注册了 %d 个工具：\n", len(definitions))
	for _, def := range definitions {
		fmt.Printf("- %s: %s\n", def.Function.Name, def.Function.Description)
	}

	// 示例：手动执行工具
	fmt.Println("\n=== 手动执行工具 ===")
	toolCall := core.ToolCall{
		ID:   "call_manual",
		Type: "function",
		Function: core.FunctionCall{
			Name:      "get_weather",
			Arguments: `{"location": "上海", "units": "celsius"}`,
		},
	}

	result, err := registry.Execute(ctx, toolCall)
	if err != nil {
		log.Printf("执行失败: %v\n", err)
	} else {
		fmt.Printf("工具执行结果: %s\n", result)
	}

	fmt.Println("\n测试完成！")
}
