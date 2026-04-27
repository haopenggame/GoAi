package test

import (
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestDocument(t *testing.T) {
	t.Run("创建文档", func(t *testing.T) {
		doc := core.NewDocument("content")
		doc.Metadata["id"] = "1"
		doc.Metadata["url"] = "https://example.com"
		assert.NotNil(t, doc)
		assert.Equal(t, "content", doc.Content)
		assert.Equal(t, "1", doc.Metadata["id"])
		assert.Equal(t, "https://example.com", doc.Metadata["url"])
	})

	t.Run("创建空文档", func(t *testing.T) {
		doc := core.NewDocument("")
		assert.NotNil(t, doc)
		assert.Equal(t, "", doc.Content)
		assert.NotNil(t, doc.Metadata)
		assert.Len(t, doc.Metadata, 0)
	})
}
