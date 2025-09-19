package domain

import (
	"context"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	// ProcessMessage handles incoming message processing
	ProcessMessage(ctx context.Context, message *Message) error
}

// ClaudeService defines the interface for Claude API interactions
type ClaudeService interface {
	// SendRequest sends a request to Claude API with context
	SendRequest(ctx context.Context, content string, contextID string) (string, error)
	
	// SendRequestWithContext sends a request using existing context
	SendRequestWithContext(ctx context.Context, content string, processingContext *ProcessingContext) (string, error)
}

// QueueConsumer defines the interface for message queue consumption
type QueueConsumer interface {
	// StartConsuming begins consuming messages from the queue
	StartConsuming(ctx context.Context, handler MessageHandler) error
	
	// StopConsuming stops message consumption
	StopConsuming() error
}

// MessageHandler defines the function signature for handling messages
type MessageHandler func(ctx context.Context, message *Message) error

// RequestService defines business logic for request management
type RequestService interface {
	// CreateRequest creates a new request from a message
	CreateRequest(ctx context.Context, message *Message) (*Request, error)
	
	// ProcessRequest handles the full request processing lifecycle
	ProcessRequest(ctx context.Context, request *Request) error
	
	// CancelRequest cancels a pending or processing request
	CancelRequest(ctx context.Context, requestID string) error
	
	// GetRequestStatus returns the current status of a request
	GetRequestStatus(ctx context.Context, requestID string) (RequestStatus, error)
	
	// GetRequestsByContext returns all requests for a given context
	GetRequestsByContext(ctx context.Context, contextID string) ([]*Request, error)
}

// ContextService defines business logic for context management
type ContextService interface {
	// GetOrCreateContext retrieves existing context or creates new one
	GetOrCreateContext(ctx context.Context, contextID string) (*ProcessingContext, error)
	
	// UpdateContext updates context with new messages
	UpdateContext(ctx context.Context, contextID string, userMessage, assistantMessage string) error
	
	// CleanupExpiredContexts removes old unused contexts
	CleanupExpiredContexts(ctx context.Context) error
}