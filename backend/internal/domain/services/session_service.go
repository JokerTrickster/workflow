package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"time"
)

// SessionService handles business logic for sessions
type SessionService struct {
	sessionRepo repositories.SessionRepository
}

// NewSessionService creates a new SessionService
func NewSessionService(sessionRepo repositories.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSession creates a new session
func (s *SessionService) CreateSession(ctx context.Context, userID string) (*entities.Session, error) {
	session := entities.NewSession(userID)
	
	if !session.IsValid() {
		return nil, errors.New("invalid session data")
	}
	
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	
	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*entities.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	if session == nil {
		return nil, errors.New("session not found")
	}
	
	// Check if session has expired
	if session.IsExpired() {
		session.Expire()
		s.sessionRepo.Update(ctx, session)
		return nil, errors.New("session has expired")
	}
	
	return session, nil
}

// GetActiveSession retrieves an active session for a user
func (s *SessionService) GetActiveSession(ctx context.Context, userID string) (*entities.Session, error) {
	sessions, err := s.sessionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Find the most recent active session
	var activeSession *entities.Session
	for _, session := range sessions {
		if !session.IsExpired() {
			if activeSession == nil || session.UpdatedAt.After(activeSession.UpdatedAt) {
				activeSession = session
			}
		}
	}
	
	if activeSession == nil {
		return nil, errors.New("no active session found")
	}
	
	return activeSession, nil
}

// ExtendSession extends the expiration time of a session
func (s *SessionService) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("session not found")
	}
	
	if session.Status != entities.SessionStatusActive {
		return errors.New("session is not active")
	}
	
	session.ExtendExpiration(duration)
	
	return s.sessionRepo.Update(ctx, session)
}

// DeactivateSession deactivates a session
func (s *SessionService) DeactivateSession(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("session not found")
	}
	
	session.Deactivate()
	
	return s.sessionRepo.Update(ctx, session)
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int, error) {
	return s.sessionRepo.CleanupExpiredSessions(ctx)
}

// AddSessionMetadata adds metadata to a session
func (s *SessionService) AddSessionMetadata(ctx context.Context, sessionID string, key string, value interface{}) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("session not found")
	}
	
	session.AddMetadata(key, value)
	
	return s.sessionRepo.Update(ctx, session)
}

// ValidateSessionAccess validates if a user can access a session
func (s *SessionService) ValidateSessionAccess(ctx context.Context, sessionID, userID string) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("session not found")
	}
	
	if session.UserID != userID {
		return errors.New("access denied: session belongs to different user")
	}
	
	if session.IsExpired() {
		return errors.New("session has expired")
	}
	
	if session.Status != entities.SessionStatusActive {
		return errors.New("session is not active")
	}
	
	return nil
}