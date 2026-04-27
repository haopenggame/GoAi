package core

import (
	"context"

	"github.com/google/uuid"
)

// Document 表示带有元数据的文本内容
type Document struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
	Score    float32
}

// NewDocument 创建一个带有自动生成ID的新文档
func NewDocument(content string) Document {
	return Document{
		ID:       uuid.New().String(),
		Content:  content,
		Metadata: make(map[string]interface{}),
	}
}

// DocumentReader 定义从各种来源读取文档的接口
type DocumentReader interface {
	Read(ctx context.Context) ([]Document, error)
}

// DocumentTransformer 定义文档转换的接口
type DocumentTransformer interface {
	Transform(ctx context.Context, documents []Document) ([]Document, error)
}

// DocumentWriter 定义文档写入的接口
type DocumentWriter interface {
	Write(ctx context.Context, documents []Document) error
}

// ContentFormatter 定义内容格式化的接口
type ContentFormatter interface {
	Format(documents []Document) string
}

// DefaultContentFormatter 是 ContentFormatter 的默认实现
type DefaultContentFormatter struct{}

// Format 实现 ContentFormatter 接口
func (f DefaultContentFormatter) Format(documents []Document) string {
	if len(documents) == 0 {
		return ""
	}
	var result string
	for i, doc := range documents {
		if i > 0 {
			result += "\n\n"
		}
		result += doc.Content
	}
	return result
}
