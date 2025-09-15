package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/application/interfaces"
)

// EventService implements task event publishing
type EventService struct {
	// In a real implementation, this would have dependencies like:
	// - Message broker client (RabbitMQ, Kafka, etc.)
	// - Event store
	// - Notification service
}

// NewEventService creates a new event service
func NewEventService() interfaces.TaskEventService {
	return &EventService{}
}

// TaskEvent represents a generic task event
type TaskEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	TaskID    string                 `json:"task_id"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishTaskCreated publishes a task created event
func (s *EventService) PublishTaskCreated(ctx context.Context, task *dto.GetTaskResponse) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.created",
		TaskID:    task.TaskID,
		UserID:    task.UserID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"title":      task.Title,
			"repository": task.Repository,
			"epic":       task.Epic,
			"branch":     task.Branch,
			"status":     task.Status,
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskUpdated publishes a task updated event
func (s *EventService) PublishTaskUpdated(ctx context.Context, task *dto.GetTaskResponse) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.updated",
		TaskID:    task.TaskID,
		UserID:    task.UserID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"version":     task.Version,
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskDeleted publishes a task deleted event
func (s *EventService) PublishTaskDeleted(ctx context.Context, taskID string, userID string) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.deleted",
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"action": "deleted",
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskStatusChanged publishes a task status change event
func (s *EventService) PublishTaskStatusChanged(ctx context.Context, taskID string, oldStatus string, newStatus string) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.status_changed",
		TaskID:    taskID,
		UserID:    "", // Will be populated by the actual user context
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
		},
	}

	return s.publishEvent(ctx, event)
}

// publishEvent publishes an event to the event system
func (s *EventService) publishEvent(ctx context.Context, event TaskEvent) error {
	// For now, we'll just log the event
	// In a real implementation, this would publish to a message broker
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ Failed to marshal event: %v", err)
		return err
	}

	log.Printf("📢 Event published: %s", string(eventJSON))

	// TODO: Implement actual event publishing
	// Examples:
	// - Publish to RabbitMQ/Kafka
	// - Store in event store
	// - Send notifications
	// - Update search indexes
	// - Trigger webhooks

	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Simple timestamp-based ID for now
	// In a real implementation, use UUID or similar
	return time.Now().Format("20060102150405") + "-" + generateRandomString(6)
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Additional event publishing methods for specific scenarios

// PublishTaskProcessingStarted publishes when a task starts processing
func (s *EventService) PublishTaskProcessingStarted(ctx context.Context, taskID string, workerID string) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.processing_started",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"worker_id": workerID,
			"status":    "processing",
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskCompleted publishes when a task is completed
func (s *EventService) PublishTaskCompleted(ctx context.Context, taskID string, tokensUsed int) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.completed",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":      "completed",
			"tokens_used": tokensUsed,
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskFailed publishes when a task fails
func (s *EventService) PublishTaskFailed(ctx context.Context, taskID string, reason string, tokensUsed int) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.failed",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":      "failed",
			"reason":      reason,
			"tokens_used": tokensUsed,
		},
	}

	return s.publishEvent(ctx, event)
}

// PublishTaskResumed publishes when a task is resumed
func (s *EventService) PublishTaskResumed(ctx context.Context, taskID string, reason string) error {
	event := TaskEvent{
		EventID:   generateEventID(),
		EventType: "task.resumed",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status": "pending",
			"reason": reason,
		},
	}

	return s.publishEvent(ctx, event)
}