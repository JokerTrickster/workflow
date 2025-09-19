package entities

import (
	"time"
)

// ProcessingContext represents the context for request processing
type ProcessingContext struct {
	ID                  string                 `json:"id"`
	RequestID           string                 `json:"request_id"`
	SessionID           string                 `json:"session_id"`
	ConversationHistory []*Message             `json:"conversation_history"`
	SystemPrompt        string                 `json:"system_prompt"`
	Metadata            map[string]interface{} `json:"metadata"`
	TokenUsage          *TokenUsage            `json:"token_usage,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// NewProcessingContext creates a new processing context
func NewProcessingContext(requestID, sessionID string) *ProcessingContext {
	now := time.Now()
	return &ProcessingContext{
		ID:                  generateID(),
		RequestID:           requestID,
		SessionID:           sessionID,
		ConversationHistory: make([]*Message, 0),
		Metadata:            make(map[string]interface{}),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// AddMessage adds a message to the conversation history
func (pc *ProcessingContext) AddMessage(message *Message) {
	pc.ConversationHistory = append(pc.ConversationHistory, message)
	pc.UpdatedAt = time.Now()
}

// SetSystemPrompt sets the system prompt for the context
func (pc *ProcessingContext) SetSystemPrompt(prompt string) {
	pc.SystemPrompt = prompt
	pc.UpdatedAt = time.Now()
}

// AddMetadata adds metadata to the processing context
func (pc *ProcessingContext) AddMetadata(key string, value interface{}) {
	if pc.Metadata == nil {
		pc.Metadata = make(map[string]interface{})
	}
	pc.Metadata[key] = value
	pc.UpdatedAt = time.Now()
}

// UpdateTokenUsage updates the token usage information
func (pc *ProcessingContext) UpdateTokenUsage(inputTokens, outputTokens int) {
	if pc.TokenUsage == nil {
		pc.TokenUsage = &TokenUsage{}
	}
	pc.TokenUsage.InputTokens += inputTokens
	pc.TokenUsage.OutputTokens += outputTokens
	pc.TokenUsage.TotalTokens = pc.TokenUsage.InputTokens + pc.TokenUsage.OutputTokens
	pc.UpdatedAt = time.Now()
}

// GetMessageCount returns the number of messages in conversation history
func (pc *ProcessingContext) GetMessageCount() int {
	return len(pc.ConversationHistory)
}

// GetLatestMessage returns the most recent message
func (pc *ProcessingContext) GetLatestMessage() *Message {
	if len(pc.ConversationHistory) == 0 {
		return nil
	}
	return pc.ConversationHistory[len(pc.ConversationHistory)-1]
}

// IsValid validates the processing context entity
func (pc *ProcessingContext) IsValid() bool {
	return pc.ID != "" &&
		pc.RequestID != "" &&
		pc.SessionID != "" &&
		!pc.CreatedAt.IsZero()
}