package core

import (
	"context"
	"fmt"
	"sync"
)

// ChatMemory 定义聊天记忆管理的接口
type ChatMemory interface {
	Add(message Message) error
	Get() ([]Message, error)
	Clear() error
}

// ConversationMemory 存储完整对话历史
type ConversationMemory struct {
	messages []Message
	mu       sync.RWMutex
}

// NewConversationMemory 创建新的对话记忆
func NewConversationMemory() *ConversationMemory {
	return &ConversationMemory{
		messages: make([]Message, 0),
	}
}

// Add 实现 ChatMemory 接口
func (m *ConversationMemory) Add(message Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

// Get 实现 ChatMemory 接口
func (m *ConversationMemory) Get() ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result, nil
}

// Clear 实现 ChatMemory 接口
func (m *ConversationMemory) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]Message, 0)
	return nil
}

// WindowedConversationMemory 存储有限窗口的对话历史
type WindowedConversationMemory struct {
	messages []Message
	maxSize  int
	mu       sync.RWMutex
}

// NewWindowedConversationMemory 创建新的窗口化对话记忆
func NewWindowedConversationMemory(maxSize int) *WindowedConversationMemory {
	return &WindowedConversationMemory{
		messages: make([]Message, 0),
		maxSize:  maxSize,
	}
}

// Add 实现 ChatMemory 接口
func (m *WindowedConversationMemory) Add(message Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	if len(m.messages) > m.maxSize {
		m.messages = m.messages[len(m.messages)-m.maxSize:]
	}
	return nil
}

// Get 实现 ChatMemory 接口
func (m *WindowedConversationMemory) Get() ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result, nil
}

// Clear 实现 ChatMemory 接口
func (m *WindowedConversationMemory) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]Message, 0)
	return nil
}

// SummaryConversationMemory 使用摘要压缩对话历史
type SummaryConversationMemory struct {
	messages    []Message
	chatModel   ChatModel
	summary     string
	maxMessages int
	mu          sync.RWMutex
}

// NewSummaryConversationMemory 创建新的摘要对话记忆
func NewSummaryConversationMemory(chatModel ChatModel, maxMessages int) *SummaryConversationMemory {
	return &SummaryConversationMemory{
		messages:    make([]Message, 0),
		chatModel:   chatModel,
		maxMessages: maxMessages,
	}
}

// Add 实现 ChatMemory 接口
func (m *SummaryConversationMemory) Add(message Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)

	if len(m.messages) > m.maxMessages {
		if err := m.summarize(); err != nil {
			return fmt.Errorf("摘要生成失败: %w", err)
		}
	}
	return nil
}

// Get 实现 ChatMemory 接口
func (m *SummaryConversationMemory) Get() ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Message
	if m.summary != "" {
		result = append(result, Message{
			Role:    RoleSystem,
			Content: fmt.Sprintf("之前对话的摘要: %s", m.summary),
		})
	}
	result = append(result, m.messages...)
	return result, nil
}

// Clear 实现 ChatMemory 接口
func (m *SummaryConversationMemory) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]Message, 0)
	m.summary = ""
	return nil
}

func (m *SummaryConversationMemory) summarize() error {
	if m.chatModel == nil {
		half := len(m.messages) / 2
		m.messages = m.messages[half:]
		return nil
	}

	half := len(m.messages) / 2
	toSummarize := m.messages[:half]
	m.messages = m.messages[half:]

	var content string
	for _, msg := range toSummarize {
		content += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	summaryPrompt := Prompt{
		Messages: []Message{
			{Role: RoleSystem, Content: "请简要总结以下对话内容。"},
			{Role: RoleUser, Content: content},
		},
	}

	ctx := context.Background()
	response, err := m.chatModel.Call(ctx, summaryPrompt)
	if err != nil {
		return err
	}

	if m.summary != "" {
		m.summary = m.summary + "\n" + response.Result.Output.Content
	} else {
		m.summary = response.Result.Output.Content
	}

	return nil
}
