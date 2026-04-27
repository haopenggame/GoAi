package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/openai"
)

type Person struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Email   string   `json:"email"`
	Hobbies []string `json:"hobbies"`
}

type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("需要设置 OPENAI_API_KEY 环境变量")
	}

	ctx := context.Background()

	client := openai.NewClient(apiKey)
	chatModel := openai.NewChatModel(client)

	fmt.Println("=== 示例 1: JSON 输出解析器 ===")
	jsonParser := core.NewJSONOutputParser(nil)

	prompt := core.NewPromptBuilder().
		User("生成一个JSON对象，表示一个叫张三的人，30岁，邮箱zhangsan@example.com，爱好是阅读和徒步。").
		Build()

	response, err := chatModel.Call(ctx, prompt)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		result, err := jsonParser.Parse(response.Result.Output.Content)
		if err != nil {
			log.Printf("解析错误: %v\n", err)
		} else {
			fmt.Printf("结果: %+v\n\n", result)
		}
	}

	fmt.Println("=== 示例 2: 结构化输出解析器 ===")
	var person Person
	structuredParser := core.NewStructuredOutputParser(&person)

	prompt2 := core.NewPromptBuilder().
		System("你是一个生成结构化数据的有用助手。").
		User("生成一个叫李四的人的信息，25岁，邮箱lisi@example.com，爱好是绘画和游泳。" + structuredParser.GetFormatInstructions()).
		Build()

	response2, err := chatModel.Call(ctx, prompt2)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		_, err := structuredParser.Parse(response2.Result.Output.Content)
		if err != nil {
			log.Printf("解析错误: %v\n", err)
		} else {
			fmt.Printf("人物: %+v\n\n", person)
		}
	}

	fmt.Println("=== 示例 3: 列表输出解析器 ===")
	listParser := core.ListOutputParser{}

	prompt3 := core.NewPromptBuilder().
		User("列出5种流行的编程语言，每行一个。").
		Build()

	response3, err := chatModel.Call(ctx, prompt3)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		result3, err := listParser.Parse(response3.Result.Output.Content)
		if err != nil {
			log.Printf("解析错误: %v\n", err)
		} else {
			languages := result3.([]string)
			fmt.Printf("编程语言:\n")
			for _, lang := range languages {
				fmt.Printf("- %s\n", lang)
			}
			fmt.Println()
		}
	}

	fmt.Println("=== 示例 4: 布尔输出解析器 ===")
	boolParser := core.BooleanOutputParser{}

	prompt5 := core.NewPromptBuilder().
		User("以下陈述是否正确？'地球是平的。'只回答是或否。").
		Build()

	response5, err := chatModel.Call(ctx, prompt5)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		result5, err := boolParser.Parse(response5.Result.Output.Content)
		if err != nil {
			log.Printf("解析错误: %v\n", err)
		} else {
			fmt.Printf("地球是平的吗？ %v\n\n", result5)
		}
	}

	fmt.Println("=== 示例 5: 逗号分隔列表解析器 ===")
	commaParser := core.CommaSeparatedListOutputParser{}

	prompt6 := core.NewPromptBuilder().
		User("列出5种颜色，用逗号分隔。").
		Build()

	response6, err := chatModel.Call(ctx, prompt6)
	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		result6, err := commaParser.Parse(response6.Result.Output.Content)
		if err != nil {
			log.Printf("解析错误: %v\n", err)
		} else {
			colors := result6.([]string)
			fmt.Printf("颜色: %v\n", colors)
		}
	}
}
