package repositories

import (
	"context"
	"local-backend-server/internal/domain/entities"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *entities.Session) error
	
	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id string) (*entities.Session, error)
	
	// GetByUserID retrieves sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error)
	
	// GetActiveByUserID retrieves active sessions for a user
	GetActiveByUserID(ctx context.Context, userID string) ([]*entities.Session, error)
	
	// Update updates an existing session
	Update(ctx context.Context, session *entities.Session) error
	
	// Delete deletes a session by ID
	Delete(ctx context.Context, id string) error
	
	// GetByStatus retrieves sessions by status
	GetByStatus(ctx context.Context, status entities.SessionStatus) ([]*entities.Session, error)
	
	// GetExpiredSessions retrieves sessions that have expired
	GetExpiredSessions(ctx context.Context) ([]*entities.Session, error)
	
	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) (int, error)
	
	// CountByStatus returns the count of sessions by status
	CountByStatus(ctx context.Context, status entities.SessionStatus) (int, error)
}