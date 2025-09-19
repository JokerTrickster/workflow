package events

import (
	"time"
)

// Message event types
const (
	MessageCreatedEventType   = "message.created"
	MessageUpdatedEventType   = "message.updated"
	MessageDeletedEventType   = "message.deleted"
	MessageMetadataAddedType  = "message.metadata_added"
)

// MessageCreatedEvent represents a message creation event
type MessageCreatedEvent struct {
	BaseEvent
	MessageID   string                 `json:"message_id"`
	SessionID   string                 `json:"session_id"`
	MessageType string                 `json:"message_type"`
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewMessageCreatedEvent creates a new message created event
func NewMessageCreatedEvent(messageID, sessionID, messageType, role, content string, metadata map[string]interface{}) *MessageCreatedEvent {
	return &MessageCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        MessageCreatedEventType,
			AggregateId: messageID,
			OccurredOn:  time.Now(),
		},
		MessageID:   messageID,
		SessionID:   sessionID,
		MessageType: messageType,
		Role:        role,
		Content:     content,
		Metadata:    metadata,
	}
}

// MessageUpdatedEvent represents a message update event
type MessageUpdatedEvent struct {
	BaseEvent
	MessageID string                 `json:"message_id"`
	SessionID string                 `json:"session_id"`
	Changes   map[string]interface{} `json:"changes"`
}

// NewMessageUpdatedEvent creates a new message updated event
func NewMessageUpdatedEvent(messageID, sessionID string, changes map[string]interface{}) *MessageUpdatedEvent {
	return &MessageUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        MessageUpdatedEventType,
			AggregateId: messageID,
			OccurredOn:  time.Now(),
		},
		MessageID: messageID,
		SessionID: sessionID,
		Changes:   changes,
	}
}

// MessageDeletedEvent represents a message deletion event
type MessageDeletedEvent struct {
	BaseEvent
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
}

// NewMessageDeletedEvent creates a new message deleted event
func NewMessageDeletedEvent(messageID, sessionID string) *MessageDeletedEvent {
	return &MessageDeletedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        MessageDeletedEventType,
			AggregateId: messageID,
			OccurredOn:  time.Now(),
		},
		MessageID: messageID,
		SessionID: sessionID,
	}
}

// MessageMetadataAddedEvent represents a message metadata addition event
type MessageMetadataAddedEvent struct {
	BaseEvent
	MessageID string      `json:"message_id"`
	SessionID string      `json:"session_id"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
}

// NewMessageMetadataAddedEvent creates a new message metadata added event
func NewMessageMetadataAddedEvent(messageID, sessionID, key string, value interface{}) *MessageMetadataAddedEvent {
	return &MessageMetadataAddedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        MessageMetadataAddedType,
			AggregateId: messageID,
			OccurredOn:  time.Now(),
		},
		MessageID: messageID,
		SessionID: sessionID,
		Key:       key,
		Value:     value,
	}
}