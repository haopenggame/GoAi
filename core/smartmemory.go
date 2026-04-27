package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryShortTerm  MemoryType = "short_term"
	MemoryLongTerm   MemoryType = "long_term"
	MemoryWorking    MemoryType = "working"
	MemoryEpisodic   MemoryType = "episodic"
	MemorySemantic   MemoryType = "semantic"
	MemoryProcedural MemoryType = "procedural"
)

// MemoryPriority 记忆优先级
type MemoryPriority int

const (
	PriorityLow      MemoryPriority = 1
	PriorityNormal   MemoryPriority = 2
	PriorityHigh     MemoryPriority = 3
	PriorityCritical MemoryPriority = 4
)

// MemoryFragment 记忆片段
type MemoryFragment struct {
	ID          string
	Content     string
	Type        MemoryType
	Priority    MemoryPriority
	Tags        []string
	Associations []string
	Embedding   []float32
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	AccessedAt  time.Time
	AccessCount int
	ExpiresAt   *time.Time
	Source      string
}

// IsExpired 检查记忆是否过期
func (f *MemoryFragment) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}

// Touch 更新访问信息
func (f *MemoryFragment) Touch() {
	f.AccessedAt = time.Now()
	f.AccessCount++
}

// Relevance 计算记忆相关性分数
func (f *MemoryFragment) Relevance() float64 {
	age := time.Since(f.CreatedAt).Hours()
	accessBonus := float64(f.AccessCount) * 0.1
	priorityBonus := float64(f.Priority) * 0.5
	recencyBonus := 1.0 / (1.0 + age*0.01)
	return recencyBonus + accessBonus + priorityBonus
}

// MemoryAssociation 记忆关联
type MemoryAssociation struct {
	SourceID    string
	TargetID    string
	RelationType string
	Strength    float64
	Metadata    map[string]interface{}
	CreatedAt   time.Time
}

// MemoryQuery 记忆查询条件
type MemoryQuery struct {
	Type        MemoryType
	Tags        []string
	Keyword     string
	PriorityMin MemoryPriority
	PriorityMax MemoryPriority
	Source      string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	MaxResults  int
	MinRelevance float64
}

// MemoryStore 记忆存储接口
type MemoryStore interface {
	Store(ctx context.Context, fragment *MemoryFragment) error
	Retrieve(ctx context.Context, id string) (*MemoryFragment, error)
	Delete(ctx context.Context, id string) error
	Query(ctx context.Context, query MemoryQuery) ([]*MemoryFragment, error)
	Associate(ctx context.Context, association MemoryAssociation) error
	GetAssociations(ctx context.Context, fragmentID string) ([]MemoryAssociation, error)
	Close() error
}

// InMemoryStore 基于内存的记忆存储实现
type InMemoryStore struct {
	mu           sync.RWMutex
	fragments    map[string]*MemoryFragment
	associations map[string][]MemoryAssociation
	indexByTag   map[string]map[string]bool
	indexByType  map[MemoryType]map[string]bool
	indexBySource map[string]map[string]bool
}

// NewInMemoryStore 创建新的内存记忆存储
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		fragments:     make(map[string]*MemoryFragment),
		associations:  make(map[string][]MemoryAssociation),
		indexByTag:    make(map[string]map[string]bool),
		indexByType:   make(map[MemoryType]map[string]bool),
		indexBySource: make(map[string]map[string]bool),
	}
}

// Store 存储记忆片段
func (s *InMemoryStore) Store(ctx context.Context, fragment *MemoryFragment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fragment.ID == "" {
		return fmt.Errorf("记忆片段ID不能为空")
	}

	s.fragments[fragment.ID] = fragment

	if fragment.Type != "" {
		if s.indexByType[fragment.Type] == nil {
			s.indexByType[fragment.Type] = make(map[string]bool)
		}
		s.indexByType[fragment.Type][fragment.ID] = true
	}

	for _, tag := range fragment.Tags {
		if s.indexByTag[tag] == nil {
			s.indexByTag[tag] = make(map[string]bool)
		}
		s.indexByTag[tag][fragment.ID] = true
	}

	if fragment.Source != "" {
		if s.indexBySource[fragment.Source] == nil {
			s.indexBySource[fragment.Source] = make(map[string]bool)
		}
		s.indexBySource[fragment.Source][fragment.ID] = true
	}

	return nil
}

// Retrieve 检索记忆片段
func (s *InMemoryStore) Retrieve(ctx context.Context, id string) (*MemoryFragment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fragment, ok := s.fragments[id]
	if !ok {
		return nil, fmt.Errorf("记忆片段 %s 不存在", id)
	}

	fragment.Touch()
	return fragment, nil
}

// Delete 删除记忆片段
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fragment, ok := s.fragments[id]
	if !ok {
		return fmt.Errorf("记忆片段 %s 不存在", id)
	}

	delete(s.fragments, id)

	if fragment.Type != "" {
		delete(s.indexByType[fragment.Type], id)
	}
	for _, tag := range fragment.Tags {
		delete(s.indexByTag[tag], id)
	}
	if fragment.Source != "" {
		delete(s.indexBySource[fragment.Source], id)
	}

	delete(s.associations, id)
	for key, assocs := range s.associations {
		var filtered []MemoryAssociation
		for _, a := range assocs {
			if a.TargetID != id {
				filtered = append(filtered, a)
			}
		}
		s.associations[key] = filtered
	}

	return nil
}

// Query 查询记忆片段
func (s *InMemoryStore) Query(ctx context.Context, query MemoryQuery) ([]*MemoryFragment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	candidates := make(map[string]bool)

	if query.Type != "" {
		if ids, ok := s.indexByType[query.Type]; ok {
			for id := range ids {
				candidates[id] = true
			}
		}
	}

	if len(query.Tags) > 0 {
		tagCandidates := make(map[string]bool)
		for _, tag := range query.Tags {
			if ids, ok := s.indexByTag[tag]; ok {
				for id := range ids {
					tagCandidates[id] = true
				}
			}
		}
		if len(candidates) == 0 {
			candidates = tagCandidates
		} else {
			for id := range candidates {
				if !tagCandidates[id] {
					delete(candidates, id)
				}
			}
		}
	}

	if query.Source != "" {
		if ids, ok := s.indexBySource[query.Source]; ok {
			if len(candidates) == 0 {
				for id := range ids {
					candidates[id] = true
				}
			} else {
				for id := range candidates {
					if !ids[id] {
						delete(candidates, id)
					}
				}
			}
		}
	}

	if len(candidates) == 0 {
		for id := range s.fragments {
			candidates[id] = true
		}
	}

	var results []*MemoryFragment
	for id := range candidates {
		fragment, ok := s.fragments[id]
		if !ok {
			continue
		}

		if fragment.IsExpired() {
			continue
		}

		if query.Keyword != "" && !strings.Contains(strings.ToLower(fragment.Content), strings.ToLower(query.Keyword)) {
			continue
		}

		if query.PriorityMin > 0 && fragment.Priority < query.PriorityMin {
			continue
		}

		if query.PriorityMax > 0 && fragment.Priority > query.PriorityMax {
			continue
		}

		if query.CreatedAfter != nil && fragment.CreatedAt.Before(*query.CreatedAfter) {
			continue
		}

		if query.CreatedBefore != nil && fragment.CreatedAt.After(*query.CreatedBefore) {
			continue
		}

		if query.MinRelevance > 0 && fragment.Relevance() < query.MinRelevance {
			continue
		}

		results = append(results, fragment)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance() > results[j].Relevance()
	})

	if query.MaxResults > 0 && len(results) > query.MaxResults {
		results = results[:query.MaxResults]
	}

	return results, nil
}

// Associate 创建记忆关联
func (s *InMemoryStore) Associate(ctx context.Context, association MemoryAssociation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if association.SourceID == "" || association.TargetID == "" {
		return fmt.Errorf("关联的源ID和目标ID不能为空")
	}

	if _, ok := s.fragments[association.SourceID]; !ok {
		return fmt.Errorf("源记忆片段 %s 不存在", association.SourceID)
	}

	if _, ok := s.fragments[association.TargetID]; !ok {
		return fmt.Errorf("目标记忆片段 %s 不存在", association.TargetID)
	}

	if association.CreatedAt.IsZero() {
		association.CreatedAt = time.Now()
	}

	s.associations[association.SourceID] = append(s.associations[association.SourceID], association)

	reverseAssoc := MemoryAssociation{
		SourceID:     association.TargetID,
		TargetID:     association.SourceID,
		RelationType: "reverse_" + association.RelationType,
		Strength:     association.Strength,
		Metadata:     association.Metadata,
		CreatedAt:    association.CreatedAt,
	}
	s.associations[association.TargetID] = append(s.associations[association.TargetID], reverseAssoc)

	return nil
}

// GetAssociations 获取记忆关联
func (s *InMemoryStore) GetAssociations(ctx context.Context, fragmentID string) ([]MemoryAssociation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assocs, ok := s.associations[fragmentID]
	if !ok {
		return []MemoryAssociation{}, nil
	}

	result := make([]MemoryAssociation, len(assocs))
	copy(result, assocs)
	return result, nil
}

// Close 关闭存储
func (s *InMemoryStore) Close() error {
	return nil
}

// FileMemoryStore 基于文件的持久化记忆存储
type FileMemoryStore struct {
	mu       sync.RWMutex
	baseDir  string
	inner    *InMemoryStore
	dirty    bool
}

// NewFileMemoryStore 创建新的文件记忆存储
func NewFileMemoryStore(baseDir string) (*FileMemoryStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	store := &FileMemoryStore{
		baseDir: baseDir,
		inner:   NewInMemoryStore(),
		dirty:   false,
	}

	if err := store.load(); err != nil {
		return store, nil
	}

	return store, nil
}

// Store 存储记忆片段
func (s *FileMemoryStore) Store(ctx context.Context, fragment *MemoryFragment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.inner.Store(ctx, fragment); err != nil {
		return err
	}

	s.dirty = true
	return s.saveFragment(fragment)
}

// Retrieve 检索记忆片段
func (s *FileMemoryStore) Retrieve(ctx context.Context, id string) (*MemoryFragment, error) {
	return s.inner.Retrieve(ctx, id)
}

// Delete 删除记忆片段
func (s *FileMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.inner.Delete(ctx, id); err != nil {
		return err
	}

	s.dirty = true
	filePath := filepath.Join(s.baseDir, "fragments", id+".json")
	_ = os.Remove(filePath)
	return nil
}

// Query 查询记忆片段
func (s *FileMemoryStore) Query(ctx context.Context, query MemoryQuery) ([]*MemoryFragment, error) {
	return s.inner.Query(ctx, query)
}

// Associate 创建记忆关联
func (s *FileMemoryStore) Associate(ctx context.Context, association MemoryAssociation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.inner.Associate(ctx, association); err != nil {
		return err
	}

	s.dirty = true
	return s.saveAssociation(association)
}

// GetAssociations 获取记忆关联
func (s *FileMemoryStore) GetAssociations(ctx context.Context, fragmentID string) ([]MemoryAssociation, error) {
	return s.inner.GetAssociations(ctx, fragmentID)
}

// Close 关闭存储并持久化
func (s *FileMemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dirty {
		return s.saveAll()
	}
	return nil
}

func (s *FileMemoryStore) load() error {
	fragmentDir := filepath.Join(s.baseDir, "fragments")
	if err := os.MkdirAll(fragmentDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(fragmentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(fragmentDir, entry.Name()))
		if err != nil {
			continue
		}

		var fragment MemoryFragment
		if err := json.Unmarshal(data, &fragment); err != nil {
			continue
		}

		_ = s.inner.Store(context.Background(), &fragment)
	}

	assocDir := filepath.Join(s.baseDir, "associations")
	assocEntries, err := os.ReadDir(assocDir)
	if err != nil {
		return nil
	}

	for _, entry := range assocEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(assocDir, entry.Name()))
		if err != nil {
			continue
		}

		var associations []MemoryAssociation
		if err := json.Unmarshal(data, &associations); err != nil {
			continue
		}

		for _, assoc := range associations {
			_ = s.inner.Associate(context.Background(), assoc)
		}
	}

	return nil
}

func (s *FileMemoryStore) saveFragment(fragment *MemoryFragment) error {
	dir := filepath.Join(s.baseDir, "fragments")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(fragment, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, fragment.ID+".json"), data, 0644)
}

func (s *FileMemoryStore) saveAssociation(association MemoryAssociation) error {
	dir := filepath.Join(s.baseDir, "associations")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(association, "", "  ")
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s_%s.json", association.SourceID, association.TargetID)
	return os.WriteFile(filepath.Join(dir, fileName), data, 0644)
}

func (s *FileMemoryStore) saveAll() error {
	fragments, err := s.inner.Query(context.Background(), MemoryQuery{})
	if err != nil {
		return err
	}

	for _, fragment := range fragments {
		if err := s.saveFragment(fragment); err != nil {
			return err
		}
	}

	return nil
}

// MemorySystem 记忆系统，提供统一的记忆管理接口
type MemorySystem struct {
	store       MemoryStore
	embedModel  EmbeddingModel
	mu          sync.RWMutex
	consolidationRules []ConsolidationRule
}

// ConsolidationRule 记忆巩固规则
type ConsolidationRule struct {
	Name        string
	Type        MemoryType
	MinAccess   int
	MinAge      time.Duration
	TargetType  MemoryType
	Priority    MemoryPriority
}

// NewMemorySystem 创建新的记忆系统
func NewMemorySystem(store MemoryStore) *MemorySystem {
	return &MemorySystem{
		store:             store,
		consolidationRules: make([]ConsolidationRule, 0),
	}
}

// WithEmbeddingModel 设置嵌入模型
func (s *MemorySystem) WithEmbeddingModel(model EmbeddingModel) *MemorySystem {
	s.embedModel = model
	return s
}

// AddConsolidationRule 添加记忆巩固规则
func (s *MemorySystem) AddConsolidationRule(rule ConsolidationRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consolidationRules = append(s.consolidationRules, rule)
}

// Remember 记住一个记忆片段
func (s *MemorySystem) Remember(ctx context.Context, content string, memType MemoryType, options ...MemoryOption) (*MemoryFragment, error) {
	fragment := &MemoryFragment{
		Content:     content,
		Type:        memType,
		Priority:    PriorityNormal,
		Tags:        make([]string, 0),
		Associations: make([]string, 0),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 0,
	}

	for _, opt := range options {
		opt(fragment)
	}

	if fragment.ID == "" {
		return nil, fmt.Errorf("记忆片段ID不能为空")
	}

	if s.embedModel != nil {
		embedding, err := s.embedModel.EmbedDocument(ctx, content)
		if err == nil {
			fragment.Embedding = embedding
		}
	}

	if err := s.store.Store(ctx, fragment); err != nil {
		return nil, fmt.Errorf("存储记忆失败: %w", err)
	}

	return fragment, nil
}

// Recall 回忆记忆
func (s *MemorySystem) Recall(ctx context.Context, id string) (*MemoryFragment, error) {
	return s.store.Retrieve(ctx, id)
}

// Forget 忘记记忆
func (s *MemorySystem) Forget(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// Search 搜索记忆
func (s *MemorySystem) Search(ctx context.Context, query MemoryQuery) ([]*MemoryFragment, error) {
	return s.store.Query(ctx, query)
}

// RecallByKeyword 按关键词回忆
func (s *MemorySystem) RecallByKeyword(ctx context.Context, keyword string, maxResults int) ([]*MemoryFragment, error) {
	return s.store.Query(ctx, MemoryQuery{
		Keyword:    keyword,
		MaxResults: maxResults,
	})
}

// RecallByTags 按标签回忆
func (s *MemorySystem) RecallByTags(ctx context.Context, tags []string, maxResults int) ([]*MemoryFragment, error) {
	return s.store.Query(ctx, MemoryQuery{
		Tags:       tags,
		MaxResults: maxResults,
	})
}

// RecallByType 按类型回忆
func (s *MemorySystem) RecallByType(ctx context.Context, memType MemoryType, maxResults int) ([]*MemoryFragment, error) {
	return s.store.Query(ctx, MemoryQuery{
		Type:       memType,
		MaxResults: maxResults,
	})
}

// Associate 关联两个记忆
func (s *MemorySystem) Associate(ctx context.Context, sourceID, targetID string, relationType string, strength float64) error {
	return s.store.Associate(ctx, MemoryAssociation{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Strength:     strength,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	})
}

// GetRelated 获取相关记忆
func (s *MemorySystem) GetRelated(ctx context.Context, fragmentID string) ([]MemoryAssociation, error) {
	return s.store.GetAssociations(ctx, fragmentID)
}

// Consolidate 执行记忆巩固（将短期记忆转化为长期记忆等）
func (s *MemorySystem) Consolidate(ctx context.Context) error {
	s.mu.RLock()
	rules := s.consolidationRules
	s.mu.RUnlock()

	for _, rule := range rules {
		fragments, err := s.store.Query(ctx, MemoryQuery{
			Type:       rule.Type,
			MinRelevance: 0,
		})
		if err != nil {
			continue
		}

		for _, fragment := range fragments {
			if fragment.AccessCount < rule.MinAccess {
				continue
			}

			if time.Since(fragment.CreatedAt) < rule.MinAge {
				continue
			}

			fragment.Type = rule.TargetType
			fragment.Priority = rule.Priority
			_ = s.store.Store(ctx, fragment)
		}
	}

	return nil
}

// Close 关闭记忆系统
func (s *MemorySystem) Close() error {
	return s.store.Close()
}

// MemoryOption 记忆选项函数
type MemoryOption func(*MemoryFragment)

// WithMemoryID 设置记忆ID
func WithMemoryID(id string) MemoryOption {
	return func(f *MemoryFragment) {
		f.ID = id
	}
}

// WithMemoryPriority 设置记忆优先级
func WithMemoryPriority(priority MemoryPriority) MemoryOption {
	return func(f *MemoryFragment) {
		f.Priority = priority
	}
}

// WithMemoryTags 设置记忆标签
func WithMemoryTags(tags ...string) MemoryOption {
	return func(f *MemoryFragment) {
		f.Tags = tags
	}
}

// WithMemorySource 设置记忆来源
func WithMemorySource(source string) MemoryOption {
	return func(f *MemoryFragment) {
		f.Source = source
	}
}

// WithMemoryExpiry 设置记忆过期时间
func WithMemoryExpiry(d time.Duration) MemoryOption {
	return func(f *MemoryFragment) {
		t := time.Now().Add(d)
		f.ExpiresAt = &t
	}
}

// WithMemoryMetadata 设置记忆元数据
func WithMemoryMetadata(metadata map[string]interface{}) MemoryOption {
	return func(f *MemoryFragment) {
		f.Metadata = metadata
	}
}
