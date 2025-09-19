package entities

import (
	"time"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusInactive SessionStatus = "inactive"
	SessionStatusExpired  SessionStatus = "expired"
)

// Session represents a conversation session entity
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id,omitempty"`
	Status    SessionStatus          `json:"status"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
}

// NewSession creates a new session entity
func NewSession(userID string) *Session {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // Default 24 hour expiration
	
	return &Session{
		ID:        generateID(),
		UserID:    userID,
		Status:    SessionStatusActive,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: &expiresAt,
	}
}

// AddMetadata adds metadata to the session
func (s *Session) AddMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// Activate marks the session as active
func (s *Session) Activate() {
	s.Status = SessionStatusActive
	s.UpdatedAt = time.Now()
}

// Deactivate marks the session as inactive
func (s *Session) Deactivate() {
	s.Status = SessionStatusInactive
	s.UpdatedAt = time.Now()
}

// Expire marks the session as expired
func (s *Session) Expire() {
	s.Status = SessionStatusExpired
	s.UpdatedAt = time.Now()
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// ExtendExpiration extends the session expiration time
func (s *Session) ExtendExpiration(duration time.Duration) {
	newExpiration := time.Now().Add(duration)
	s.ExpiresAt = &newExpiration
	s.UpdatedAt = time.Now()
}

// IsValid validates the session entity
func (s *Session) IsValid() bool {
	return s.ID != "" &&
		s.Status != "" &&
		!s.CreatedAt.IsZero()
}