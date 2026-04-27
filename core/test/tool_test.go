package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestFunctionTool(t *testing.T) {
	t.Run("创建工具", func(t *testing.T) {
		tool := core.NewFunctionToolBuilder("get_weather").
			WithDescription("获取指定位置的天气信息").
			WithParameter("location", "string", "城市名称", true).
			WithParameter("units", "string", "温度单位", false).
			Build()

		assert.NotNil(t, tool)
		assert.Equal(t, "get_weather", tool.Function.Name)
		assert.Equal(t, "获取指定位置的天气信息", tool.Function.Description)
		properties, ok := tool.Function.Parameters["properties"].(map[string]interface{})
		assert.True(t, ok)
		assert.Len(t, properties, 2)
	})

	t.Run("创建无参数工具", func(t *testing.T) {
		tool := core.NewFunctionToolBuilder("get_time").
			WithDescription("获取当前时间").
			Build()

		assert.NotNil(t, tool)
		assert.Equal(t, "get_time", tool.Function.Name)
		assert.Equal(t, "获取当前时间", tool.Function.Description)
		properties, ok := tool.Function.Parameters["properties"].(map[string]interface{})
		assert.True(t, ok)
		assert.Len(t, properties, 0)
	})
}

func TestToolRegistry(t *testing.T) {
	t.Run("注册和调用工具", func(t *testing.T) {
		registry := core.NewToolRegistry()

		// 注册工具
		tool := core.NewFunctionToolBuilder("add").
			WithDescription("两个数相加").
			WithParameter("a", "number", "第一个数", true).
			WithParameter("b", "number", "第二个数", true).
			Build()

		err := registry.Register(tool, func(ctx context.Context, arguments string) (string, error) {
			var args struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			err := core.ParseArguments(arguments, &args)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%.2f", args.A+args.B), nil
		})
		assert.NoError(t, err)

		// 获取工具定义
		registeredTool, ok := registry.GetDefinition("add")
		assert.True(t, ok)
		assert.NotNil(t, registeredTool)

		// 调用工具
		toolCall := core.ToolCall{
			Function: core.FunctionCall{
				Name:      "add",
				Arguments: `{"a": 1, "b": 2}`,
			},
		}
		result, err := registry.Execute(context.Background(), toolCall)
		assert.NoError(t, err)
		assert.Equal(t, "3.00", result)
	})

	t.Run("调用不存在的工具", func(t *testing.T) {
		registry := core.NewToolRegistry()
		toolCall := core.ToolCall{
			Function: core.FunctionCall{
				Name:      "non_existent",
				Arguments: "{}",
			},
		}
		result, err := registry.Execute(context.Background(), toolCall)
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

func TestToolCalling(t *testing.T) {
	t.Run("解析工具调用参数", func(t *testing.T) {
		arguments := `{"location": "北京", "units": "celsius"}`
		var args struct {
			Location string `json:"location"`
			Units    string `json:"units"`
		}
		err := core.ParseArguments(arguments, &args)
		assert.NoError(t, err)
		assert.Equal(t, "北京", args.Location)
		assert.Equal(t, "celsius", args.Units)
	})

	t.Run("解析空参数", func(t *testing.T) {
		arguments := `{}`
		var args struct {
			Location string `json:"location"`
		}
		err := core.ParseArguments(arguments, &args)
		assert.NoError(t, err)
		assert.Empty(t, args.Location)
	})
}