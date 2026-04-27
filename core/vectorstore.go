package core

import "context"

// VectorStore 定义向量数据库操作的接口
type VectorStore interface {
	Add(ctx context.Context, documents []Document) error
	Delete(ctx context.Context, ids []string) error
	SimilaritySearch(ctx context.Context, query string, options SearchOptions) ([]Document, error)
}

// SearchOptions 包含相似性搜索的选项
type SearchOptions struct {
	TopK                int
	SimilarityThreshold float32
	Filter              FilterExpression
}

// FilterExpression 表示向量搜索的过滤表达式
type FilterExpression struct {
	Expression string
	Parameters map[string]interface{}
}

// DocumentRetriever 定义文档检索的接口
type DocumentRetriever interface {
	Retrieve(ctx context.Context, query string, options RetrieveOptions) ([]Document, error)
}

// RetrieveOptions 包含文档检索的选项
type RetrieveOptions struct {
	TopK   int
	Filter FilterExpression
}

// SimpleVectorStore 是简单的内存向量存储实现
type SimpleVectorStore struct {
	documents  []Document
	embeddings map[string][]float32
	model      EmbeddingModel
}

// NewSimpleVectorStore 创建新的简单向量存储
func NewSimpleVectorStore(model EmbeddingModel) *SimpleVectorStore {
	return &SimpleVectorStore{
		documents:  make([]Document, 0),
		embeddings: make(map[string][]float32),
		model:      model,
	}
}

// Add 实现 VectorStore 接口
func (s *SimpleVectorStore) Add(ctx context.Context, documents []Document) error {
	for _, doc := range documents {
		embedding, err := s.model.EmbedDocument(ctx, doc.Content)
		if err != nil {
			return err
		}
		s.documents = append(s.documents, doc)
		s.embeddings[doc.ID] = embedding
	}
	return nil
}

// Delete 实现 VectorStore 接口
func (s *SimpleVectorStore) Delete(ctx context.Context, ids []string) error {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	var filtered []Document
	for _, doc := range s.documents {
		if !idSet[doc.ID] {
			filtered = append(filtered, doc)
		} else {
			delete(s.embeddings, doc.ID)
		}
	}
	s.documents = filtered
	return nil
}

// SimilaritySearch 实现 VectorStore 接口
func (s *SimpleVectorStore) SimilaritySearch(ctx context.Context, query string, options SearchOptions) ([]Document, error) {
	queryEmbedding, err := s.model.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	type scoredDoc struct {
		doc   Document
		score float32
	}

	var scored []scoredDoc
	for _, doc := range s.documents {
		embedding, ok := s.embeddings[doc.ID]
		if !ok {
			continue
		}
		similarity := cosineSimilarity(queryEmbedding, embedding)
		if similarity >= options.SimilarityThreshold {
			scored = append(scored, scoredDoc{doc: doc, score: similarity})
		}
	}

	// 按分数降序排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	var results []Document
	for i := 0; i < len(scored) && i < options.TopK; i++ {
		doc := scored[i].doc
		doc.Score = scored[i].score
		results = append(results, doc)
	}
	return results, nil
}

// cosineSimilarity 计算两个向量之间的余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float32
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (normA * normB)
}
