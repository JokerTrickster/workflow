package domain

import (
	"time"
)

// MessageType represents the type of message received from RabbitMQ
type MessageType string

const (
	MessageTypeWorkRequest    MessageType = "work_request"
	MessageTypeCancellation   MessageType = "work_cancellation"
)

// RequestStatus represents the current status of a processing request
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
	RequestStatusCancelled  RequestStatus = "cancelled"
)

// Message represents a message received from RabbitMQ
type Message struct {
	ID          string                 `json:"id"`
	Type        MessageType            `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	ContextID   *string                `json:"context_id,omitempty"`
	ReceivedAt  time.Time              `json:"received_at"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// NewMessage creates a new Message instance
func NewMessage(id string, msgType MessageType, payload map[string]interface{}) *Message {
	return &Message{
		ID:         id,
		Type:       msgType,
		Payload:    payload,
		ReceivedAt: time.Now(),
		Metadata:   make(map[string]string),
	}
}

// SetContextID sets the context ID for conversation management
func (m *Message) SetContextID(contextID string) {
	m.ContextID = &contextID
}

// GetContextID returns the context ID if set
func (m *Message) GetContextID() string {
	if m.ContextID == nil {
		return ""
	}
	return *m.ContextID
}

// IsWorkRequest checks if this is a work request message
func (m *Message) IsWorkRequest() bool {
	return m.Type == MessageTypeWorkRequest
}

// IsCancellation checks if this is a cancellation message
func (m *Message) IsCancellation() bool {
	return m.Type == MessageTypeCancellation
}

// Request represents a processing request with status tracking
type Request struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	MessageID     string        `json:"message_id" gorm:"index"`
	ContextID     string        `json:"context_id" gorm:"index"`
	Status        RequestStatus `json:"status" gorm:"index"`
	RequestData   string        `json:"request_data" gorm:"type:text"` // JSON serialized
	Response      string        `json:"response" gorm:"type:text"`
	ErrorMessage  string        `json:"error_message" gorm:"type:text"`
	StartedAt     *time.Time    `json:"started_at"`
	CompletedAt   *time.Time    `json:"completed_at"`
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewRequest creates a new Request instance
func NewRequest(id, messageID, contextID string, requestData string) *Request {
	return &Request{
		ID:          id,
		MessageID:   messageID,
		ContextID:   contextID,
		Status:      RequestStatusPending,
		RequestData: requestData,
	}
}

// Start marks the request as processing
func (r *Request) Start() {
	r.Status = RequestStatusProcessing
	now := time.Now()
	r.StartedAt = &now
}

// Complete marks the request as completed with response
func (r *Request) Complete(response string) {
	r.Status = RequestStatusCompleted
	r.Response = response
	now := time.Now()
	r.CompletedAt = &now
}

// Fail marks the request as failed with error message
func (r *Request) Fail(errorMessage string) {
	r.Status = RequestStatusFailed
	r.ErrorMessage = errorMessage
	now := time.Now()
	r.CompletedAt = &now
}

// Cancel marks the request as cancelled
func (r *Request) Cancel() {
	r.Status = RequestStatusCancelled
	now := time.Now()
	r.CompletedAt = &now
}

// IsCompleted checks if the request is in a final state
func (r *Request) IsCompleted() bool {
	return r.Status == RequestStatusCompleted ||
		   r.Status == RequestStatusFailed ||
		   r.Status == RequestStatusCancelled
}

// CanBeCancelled checks if the request can be cancelled
func (r *Request) CanBeCancelled() bool {
	return r.Status == RequestStatusPending || r.Status == RequestStatusProcessing
}

// ProcessingContext holds context for Claude API conversations
type ProcessingContext struct {
	ID            string            `json:"id" gorm:"primaryKey"`
	Messages      []ContextMessage  `json:"messages" gorm:"serializer:json"`
	Metadata      map[string]string `json:"metadata" gorm:"serializer:json"`
	CreatedAt     time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	LastUsedAt    time.Time         `json:"last_used_at"`
}

// ContextMessage represents a message in the conversation context
type ContextMessage struct {
	Role      string    `json:"role"`      // "user", "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewProcessingContext creates a new ProcessingContext
func NewProcessingContext(id string) *ProcessingContext {
	now := time.Now()
	return &ProcessingContext{
		ID:         id,
		Messages:   make([]ContextMessage, 0),
		Metadata:   make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
}

// AddUserMessage adds a user message to the context
func (pc *ProcessingContext) AddUserMessage(content string) {
	pc.Messages = append(pc.Messages, ContextMessage{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})
	pc.LastUsedAt = time.Now()
}

// AddAssistantMessage adds an assistant message to the context
func (pc *ProcessingContext) AddAssistantMessage(content string) {
	pc.Messages = append(pc.Messages, ContextMessage{
		Role:      "assistant", 
		Content:   content,
		Timestamp: time.Now(),
	})
	pc.LastUsedAt = time.Now()
}

// GetMessages returns all messages in the context
func (pc *ProcessingContext) GetMessages() []ContextMessage {
	return pc.Messages
}

// IsExpired checks if context has expired based on last usage
func (pc *ProcessingContext) IsExpired(maxAge time.Duration) bool {
	return time.Since(pc.LastUsedAt) > maxAge
}

// GetMetadata returns metadata value by key
func (pc *ProcessingContext) GetMetadata(key string) (string, bool) {
	value, exists := pc.Metadata[key]
	return value, exists
}

// SetMetadata sets a metadata key-value pair
func (pc *ProcessingContext) SetMetadata(key, value string) {
	if pc.Metadata == nil {
		pc.Metadata = make(map[string]string)
	}
	pc.Metadata[key] = value
}