package test

import (
	"context"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestVectorStore(t *testing.T) {
	t.Run("创建向量存储", func(t *testing.T) {
		model := &MockEmbeddingModel{embedding: make([]float32, 3)}
		store := core.NewSimpleVectorStore(model)
		assert.NotNil(t, store)
	})

	t.Run("添加和检索文档", func(t *testing.T) {
		ctx := context.Background()
		model := &MockEmbeddingModel{embedding: make([]float32, 3)}
		store := core.NewSimpleVectorStore(model)

		// 添加文档
		doc1 := core.NewDocument("Hello world")
		doc1.Metadata["id"] = "1"
		doc2 := core.NewDocument("Hello Go")
		doc2.Metadata["id"] = "2"

		err := store.Add(ctx, []core.Document{doc1, doc2})
		assert.NoError(t, err)

		// 检索文档
		results, err := store.SimilaritySearch(ctx, "Hello world", core.SearchOptions{TopK: 1, SimilarityThreshold: 0})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Hello world", results[0].Content)
	})

	t.Run("空存储检索", func(t *testing.T) {
		ctx := context.Background()
		model := &MockEmbeddingModel{embedding: make([]float32, 3)}
		store := core.NewSimpleVectorStore(model)

		results, err := store.SimilaritySearch(ctx, "Hello world", core.SearchOptions{TopK: 1, SimilarityThreshold: 0})
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestSearchOptions(t *testing.T) {
	t.Run("创建搜索选项", func(t *testing.T) {
		options := core.SearchOptions{
			TopK:                5,
			SimilarityThreshold: 0.8,
		}
		assert.Equal(t, 5, options.TopK)
		assert.Equal(t, float32(0.8), options.SimilarityThreshold)
	})
}
