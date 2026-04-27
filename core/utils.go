package core

import (
	"context"
	"fmt"
	"strings"
)

// NumberedContentFormatter 带编号的内容格式化器
type NumberedContentFormatter struct{}

// Format 实现 ContentFormatter 接口
func (f NumberedContentFormatter) Format(documents []Document) string {
	var parts []string
	for i, doc := range documents {
		parts = append(parts, fmt.Sprintf("[%d] %s", i+1, doc.Content))
	}
	return strings.Join(parts, "\n\n")
}

// VectorStoreWriter 将文档写入向量存储
type VectorStoreWriter struct {
	Store VectorStore
}

// Write 实现 DocumentWriter 接口
func (w *VectorStoreWriter) Write(ctx context.Context, documents []Document) error {
	return w.Store.Add(ctx, documents)
}
