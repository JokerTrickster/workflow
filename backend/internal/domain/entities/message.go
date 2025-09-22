package entities

import (
	"time"
)

// MessageType represents different types of messages
type MessageType string

const (
	MessageTypeWorkRequest MessageType = "work_request"
	MessageTypeCancel      MessageType = "cancel"
	MessageTypeStatus      MessageType = "status"
	MessageTypeClaudeTask  MessageType = "claude_task"
)

// MessageRole represents the role of the message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// Message represents a chat message entity
type Message struct {
	ID        string      `json:"id"`
	SessionID string      `json:"session_id"`
	Type      MessageType `json:"type"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	Metadata  Metadata    `json:"metadata"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Metadata holds additional message information
type Metadata map[string]interface{}

// NewMessage creates a new message entity
func NewMessage(sessionID string, messageType MessageType, role MessageRole, content string) *Message {
	now := time.Now()
	return &Message{
		ID:        generateID(),
		SessionID: sessionID,
		Type:      messageType,
		Role:      role,
		Content:   content,
		Metadata:  make(Metadata),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMetadata adds metadata to the message
func (m *Message) AddMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(Metadata)
	}
	m.Metadata[key] = value
	m.UpdatedAt = time.Now()
}

// GetMetadata retrieves metadata by key
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}

// IsValid validates the message entity
func (m *Message) IsValid() bool {
	return m.ID != "" &&
		m.SessionID != "" &&
		m.Type != "" &&
		m.Role != "" &&
		m.Content != "" &&
		!m.CreatedAt.IsZero()
}