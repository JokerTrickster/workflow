package events

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Request event types
const (
	RequestCreatedEventType   = "request.created"
	RequestStartedEventType   = "request.started"
	RequestCompletedEventType = "request.completed"
	RequestFailedEventType    = "request.failed"
	RequestCancelledEventType = "request.cancelled"
	RequestTimeoutEventType   = "request.timeout"
)

// RequestCreatedEvent represents a request creation event
type RequestCreatedEvent struct {
	BaseEvent
	RequestID   string                 `json:"request_id"`
	SessionID   string                 `json:"session_id"`
	RequestType string                 `json:"request_type"`
	Input       map[string]interface{} `json:"input"`
}

// NewRequestCreatedEvent creates a new request created event
func NewRequestCreatedEvent(requestID, sessionID, requestType string, input map[string]interface{}) *RequestCreatedEvent {
	return &RequestCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestCreatedEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID:   requestID,
		SessionID:   sessionID,
		RequestType: requestType,
		Input:       input,
	}
}

// RequestStartedEvent represents a request start event
type RequestStartedEvent struct {
	BaseEvent
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
}

// NewRequestStartedEvent creates a new request started event
func NewRequestStartedEvent(requestID, sessionID string) *RequestStartedEvent {
	return &RequestStartedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestStartedEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID: requestID,
		SessionID: sessionID,
	}
}

// RequestCompletedEvent represents a request completion event
type RequestCompletedEvent struct {
	BaseEvent
	RequestID        string                 `json:"request_id"`
	SessionID        string                 `json:"session_id"`
	Output           map[string]interface{} `json:"output"`
	ProcessingTimeMs int64                  `json:"processing_time_ms"`
}

// NewRequestCompletedEvent creates a new request completed event
func NewRequestCompletedEvent(requestID, sessionID string, output map[string]interface{}, processingTimeMs int64) *RequestCompletedEvent {
	return &RequestCompletedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestCompletedEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID:        requestID,
		SessionID:        sessionID,
		Output:           output,
		ProcessingTimeMs: processingTimeMs,
	}
}

// RequestFailedEvent represents a request failure event
type RequestFailedEvent struct {
	BaseEvent
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	Error     string `json:"error"`
}

// NewRequestFailedEvent creates a new request failed event
func NewRequestFailedEvent(requestID, sessionID, errorMsg string) *RequestFailedEvent {
	return &RequestFailedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestFailedEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID: requestID,
		SessionID: sessionID,
		Error:     errorMsg,
	}
}

// RequestCancelledEvent represents a request cancellation event
type RequestCancelledEvent struct {
	BaseEvent
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
}

// NewRequestCancelledEvent creates a new request cancelled event
func NewRequestCancelledEvent(requestID, sessionID string) *RequestCancelledEvent {
	return &RequestCancelledEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestCancelledEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID: requestID,
		SessionID: sessionID,
	}
}

// RequestTimeoutEvent represents a request timeout event
type RequestTimeoutEvent struct {
	BaseEvent
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
}

// NewRequestTimeoutEvent creates a new request timeout event
func NewRequestTimeoutEvent(requestID, sessionID string) *RequestTimeoutEvent {
	return &RequestTimeoutEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        RequestTimeoutEventType,
			AggregateId: requestID,
			OccurredOn:  time.Now(),
		},
		RequestID: requestID,
		SessionID: sessionID,
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}