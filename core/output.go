package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// OutputParser 定义解析模型输出的接口
type OutputParser interface {
	Parse(output string) (interface{}, error)
	GetFormatInstructions() string
}

// StringOutputParser 返回原始字符串输出
type StringOutputParser struct{}

// Parse 实现 OutputParser 接口
func (p StringOutputParser) Parse(output string) (interface{}, error) {
	return output, nil
}

// GetFormatInstructions 实现 OutputParser 接口
func (p StringOutputParser) GetFormatInstructions() string {
	return ""
}

// JSONOutputParser 将输出解析为 JSON
type JSONOutputParser struct {
	Schema interface{}
}

// NewJSONOutputParser 创建新的 JSON 输出解析器
func NewJSONOutputParser(schema interface{}) JSONOutputParser {
	return JSONOutputParser{Schema: schema}
}

// Parse 实现 OutputParser 接口
func (p JSONOutputParser) Parse(output string) (interface{}, error) {
	output = extractJSON(output)
	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}
	return result, nil
}

// GetFormatInstructions 实现 OutputParser 接口
func (p JSONOutputParser) GetFormatInstructions() string {
	return "请以JSON格式返回结果。"
}

// ListOutputParser 将输出解析为字符串列表
type ListOutputParser struct{}

// Parse 实现 OutputParser 接口
func (p ListOutputParser) Parse(output string) (interface{}, error) {
	lines := strings.Split(output, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 移除列表编号前缀
		line = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(line, "")
		line = regexp.MustCompile(`^[-*]\s*`).ReplaceAllString(line, "")
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

// GetFormatInstructions 实现 OutputParser 接口
func (p ListOutputParser) GetFormatInstructions() string {
	return "请以列表格式返回结果，每行一个项目。"
}

// StructuredOutputParser 将输出解析到结构体
type StructuredOutputParser struct {
	Schema interface{}
}

// NewStructuredOutputParser 创建新的结构化输出解析器
func NewStructuredOutputParser(schema interface{}) StructuredOutputParser {
	return StructuredOutputParser{Schema: schema}
}

// Parse 实现 OutputParser 接口
func (p StructuredOutputParser) Parse(output string) (interface{}, error) {
	output = extractJSON(output)
	if err := json.Unmarshal([]byte(output), p.Schema); err != nil {
		return nil, fmt.Errorf("结构化解析失败: %w", err)
	}
	return p.Schema, nil
}

// GetFormatInstructions 实现 OutputParser 接口
func (p StructuredOutputParser) GetFormatInstructions() string {
	schemaBytes, err := json.MarshalIndent(p.Schema, "", "  ")
	if err != nil {
		return "请以JSON格式返回结果。"
	}
	return fmt.Sprintf("请以以下JSON格式返回结果:\n```json\n%s\n```", string(schemaBytes))
}

// BooleanOutputParser 将输出解析为布尔值
type BooleanOutputParser struct{}

// Parse 实现 OutputParser 接口
func (p BooleanOutputParser) Parse(output string) (interface{}, error) {
	output = strings.TrimSpace(strings.ToLower(output))
	if output == "true" || output == "yes" || output == "是" || output == "1" {
		return true, nil
	}
	if output == "false" || output == "no" || output == "否" || output == "0" {
		return false, nil
	}
	return nil, fmt.Errorf("无法将 '%s' 解析为布尔值", output)
}

// GetFormatInstructions 实现 OutputParser 接口
func (p BooleanOutputParser) GetFormatInstructions() string {
	return "请回答是或否。"
}

// CommaSeparatedListOutputParser 将逗号分隔的输出解析为列表
type CommaSeparatedListOutputParser struct{}

// Parse 实现 OutputParser 接口
func (p CommaSeparatedListOutputParser) Parse(output string) (interface{}, error) {
	items := strings.Split(output, ",")
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result, nil
}

// GetFormatInstructions 实现 OutputParser 接口
func (p CommaSeparatedListOutputParser) GetFormatInstructions() string {
	return "请以逗号分隔的列表返回结果。"
}

// extractJSON 从可能包含 markdown 代码块的文本中提取 JSON
func extractJSON(text string) string {
	// 尝试提取 markdown 代码块中的 JSON
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(text)
}
