package events

import (
	"time"
)

// Event represents a domain event
type Event interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	Payload() interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	AggregateId string      `json:"aggregate_id"`
	OccurredOn  time.Time   `json:"occurred_at"`
	Data        interface{} `json:"payload"`
}

// EventID returns the event ID
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID
func (e BaseEvent) AggregateID() string {
	return e.AggregateId
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.OccurredOn
}

// Payload returns the event payload
func (e BaseEvent) Payload() interface{} {
	return e.Data
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(event Event) error
	PublishBatch(events []Event) error
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	Handle(event Event) error
	CanHandle(eventType string) bool
}