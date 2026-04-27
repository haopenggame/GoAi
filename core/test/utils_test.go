package test

import (
	"context"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestNumberedContentFormatter(t *testing.T) {
	formatter := core.NumberedContentFormatter{}

	docs := []core.Document{
		{Content: "第一段内容"},
		{Content: "第二段内容"},
		{Content: "第三段内容"},
	}

	result := formatter.Format(docs)
	assert.Contains(t, result, "[1] 第一段内容")
	assert.Contains(t, result, "[2] 第二段内容")
	assert.Contains(t, result, "[3] 第三段内容")
}

func TestVectorStoreWriter(t *testing.T) {
	ctx := context.Background()
	model := &MockEmbeddingModel{embedding: make([]float32, 10)}
	store := core.NewSimpleVectorStore(model)
	writer := &core.VectorStoreWriter{Store: store}

	docs := []core.Document{
		core.NewDocument("文档1"),
		core.NewDocument("文档2"),
		core.NewDocument("文档3"),
	}

	err := writer.Write(ctx, docs)
	assert.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "文档", core.SearchOptions{TopK: 10, SimilarityThreshold: 0})
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}
