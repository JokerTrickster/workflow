package repositories

import (
	"context"
	"local-backend-server/internal/domain/entities"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *entities.Message) error
	
	// GetByID retrieves a message by its ID
	GetByID(ctx context.Context, id string) (*entities.Message, error)
	
	// GetBySessionID retrieves all messages for a session
	GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Message, error)
	
	// GetBySessionIDWithPagination retrieves messages with pagination
	GetBySessionIDWithPagination(ctx context.Context, sessionID string, offset, limit int) ([]*entities.Message, error)
	
	// Update updates an existing message
	Update(ctx context.Context, message *entities.Message) error
	
	// Delete deletes a message by ID
	Delete(ctx context.Context, id string) error
	
	// GetByTypeAndSessionID retrieves messages by type and session
	GetByTypeAndSessionID(ctx context.Context, messageType entities.MessageType, sessionID string) ([]*entities.Message, error)
	
	// CountBySessionID returns the count of messages in a session
	CountBySessionID(ctx context.Context, sessionID string) (int, error)
	
	// GetLatestBySessionID retrieves the most recent message in a session
	GetLatestBySessionID(ctx context.Context, sessionID string) (*entities.Message, error)
}