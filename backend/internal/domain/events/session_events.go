package events

import (
	"time"
)

// Session event types
const (
	SessionCreatedEventType    = "session.created"
	SessionActivatedEventType  = "session.activated"
	SessionDeactivatedEventType = "session.deactivated"
	SessionExpiredEventType    = "session.expired"
	SessionExtendedEventType   = "session.extended"
)

// SessionCreatedEvent represents a session creation event
type SessionCreatedEvent struct {
	BaseEvent
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	ExpiresAt *time.Time             `json:"expires_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewSessionCreatedEvent creates a new session created event
func NewSessionCreatedEvent(sessionID, userID string, expiresAt *time.Time, metadata map[string]interface{}) *SessionCreatedEvent {
	return &SessionCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        SessionCreatedEventType,
			AggregateId: sessionID,
			OccurredOn:  time.Now(),
		},
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Metadata:  metadata,
	}
}

// SessionActivatedEvent represents a session activation event
type SessionActivatedEvent struct {
	BaseEvent
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

// NewSessionActivatedEvent creates a new session activated event
func NewSessionActivatedEvent(sessionID, userID string) *SessionActivatedEvent {
	return &SessionActivatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        SessionActivatedEventType,
			AggregateId: sessionID,
			OccurredOn:  time.Now(),
		},
		SessionID: sessionID,
		UserID:    userID,
	}
}

// SessionDeactivatedEvent represents a session deactivation event
type SessionDeactivatedEvent struct {
	BaseEvent
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

// NewSessionDeactivatedEvent creates a new session deactivated event
func NewSessionDeactivatedEvent(sessionID, userID string) *SessionDeactivatedEvent {
	return &SessionDeactivatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        SessionDeactivatedEventType,
			AggregateId: sessionID,
			OccurredOn:  time.Now(),
		},
		SessionID: sessionID,
		UserID:    userID,
	}
}

// SessionExpiredEvent represents a session expiration event
type SessionExpiredEvent struct {
	BaseEvent
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

// NewSessionExpiredEvent creates a new session expired event
func NewSessionExpiredEvent(sessionID, userID string) *SessionExpiredEvent {
	return &SessionExpiredEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        SessionExpiredEventType,
			AggregateId: sessionID,
			OccurredOn:  time.Now(),
		},
		SessionID: sessionID,
		UserID:    userID,
	}
}

// SessionExtendedEvent represents a session extension event
type SessionExtendedEvent struct {
	BaseEvent
	SessionID     string        `json:"session_id"`
	UserID        string        `json:"user_id"`
	ExtensionTime time.Duration `json:"extension_time"`
	NewExpiresAt  time.Time     `json:"new_expires_at"`
}

// NewSessionExtendedEvent creates a new session extended event
func NewSessionExtendedEvent(sessionID, userID string, extensionTime time.Duration, newExpiresAt time.Time) *SessionExtendedEvent {
	return &SessionExtendedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        SessionExtendedEventType,
			AggregateId: sessionID,
			OccurredOn:  time.Now(),
		},
		SessionID:     sessionID,
		UserID:        userID,
		ExtensionTime: extensionTime,
		NewExpiresAt:  newExpiresAt,
	}
}