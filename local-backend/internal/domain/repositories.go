package domain

import (
	"context"
	"time"
)

// RequestRepository defines the interface for request data access
type RequestRepository interface {
	// Create saves a new request
	Create(ctx context.Context, request *Request) error
	
	// GetByID retrieves a request by its ID
	GetByID(ctx context.Context, id string) (*Request, error)
	
	// GetByMessageID retrieves a request by message ID
	GetByMessageID(ctx context.Context, messageID string) (*Request, error)
	
	// GetByContextID retrieves all requests for a context
	GetByContextID(ctx context.Context, contextID string) ([]*Request, error)
	
	// GetByStatus retrieves requests by status
	GetByStatus(ctx context.Context, status RequestStatus) ([]*Request, error)
	
	// Update updates an existing request
	Update(ctx context.Context, request *Request) error
	
	// Delete removes a request (soft delete recommended)
	Delete(ctx context.Context, id string) error
	
	// GetPendingRequests retrieves all pending requests
	GetPendingRequests(ctx context.Context) ([]*Request, error)
	
	// GetProcessingRequests retrieves all currently processing requests
	GetProcessingRequests(ctx context.Context) ([]*Request, error)
	
	// CountByStatus counts requests by status
	CountByStatus(ctx context.Context, status RequestStatus) (int64, error)
	
	// GetRequestsCreatedBetween retrieves requests created within a time range
	GetRequestsCreatedBetween(ctx context.Context, start, end time.Time) ([]*Request, error)
}

// ProcessingContextRepository defines the interface for context data access
type ProcessingContextRepository interface {
	// Create saves a new processing context
	Create(ctx context.Context, context *ProcessingContext) error
	
	// GetByID retrieves a context by its ID
	GetByID(ctx context.Context, id string) (*ProcessingContext, error)
	
	// Update updates an existing context
	Update(ctx context.Context, context *ProcessingContext) error
	
	// Delete removes a context
	Delete(ctx context.Context, id string) error
	
	// GetExpiredContexts retrieves contexts that haven't been used recently
	GetExpiredContexts(ctx context.Context, maxAge int64) ([]*ProcessingContext, error)
	
	// CleanupExpiredContexts removes old contexts
	CleanupExpiredContexts(ctx context.Context, maxAge int64) error
	
	// GetContextStats returns statistics about contexts
	GetContextStats(ctx context.Context) (map[string]interface{}, error)
	
	// GetContextsCreatedAfter retrieves contexts created after a specific time
	GetContextsCreatedAfter(ctx context.Context, since time.Time) ([]*ProcessingContext, error)
}