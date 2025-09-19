package repositories

import (
	"context"
	"local-backend-server/internal/domain/entities"
)

// ProcessingContextRepository defines the interface for processing context data access
type ProcessingContextRepository interface {
	// Create creates a new processing context
	Create(ctx context.Context, context *entities.ProcessingContext) error
	
	// GetByID retrieves a processing context by its ID
	GetByID(ctx context.Context, id string) (*entities.ProcessingContext, error)
	
	// GetByRequestID retrieves a processing context by request ID
	GetByRequestID(ctx context.Context, requestID string) (*entities.ProcessingContext, error)
	
	// GetBySessionID retrieves processing contexts for a session
	GetBySessionID(ctx context.Context, sessionID string) ([]*entities.ProcessingContext, error)
	
	// Update updates an existing processing context
	Update(ctx context.Context, context *entities.ProcessingContext) error
	
	// Delete deletes a processing context by ID
	Delete(ctx context.Context, id string) error
	
	// GetLatestBySessionID retrieves the most recent processing context for a session
	GetLatestBySessionID(ctx context.Context, sessionID string) (*entities.ProcessingContext, error)
	
	// GetContextsWithMessages retrieves processing contexts with their conversation history
	GetContextsWithMessages(ctx context.Context, sessionID string) ([]*entities.ProcessingContext, error)
}