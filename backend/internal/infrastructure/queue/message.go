package queue

import (
	"encoding/json"
	"time"

	"ai-git-workbench/internal/domain/entities"
)

// MessageType defines the types of queue messages
type MessageType string

const (
	MessageTypeTaskCreated MessageType = "task_created"
	MessageTypeTaskUpdated MessageType = "task_updated"
	MessageTypeTaskResumed MessageType = "task_resumed"
)

// QueueMessage represents a message in the RabbitMQ queue
type QueueMessage struct {
	ID         string      `json:"id"`
	Type       MessageType `json:"type"`
	Payload    TaskPayload `json:"payload"`
	Timestamp  time.Time   `json:"timestamp"`
	RetryCount int         `json:"retry_count"`
}

// TaskPayload contains the task data in the message
type TaskPayload struct {
	TaskID     string            `json:"task_id"`
	UserID     string            `json:"user_id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Repository string            `json:"repository"`
	BranchName string            `json:"branch_name"`
	Status     string            `json:"status"`
	Epic       string            `json:"epic"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	StartedAt  *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	TokensUsed int               `json:"tokens_used"`
	Version    int64             `json:"version"`
}

// NewTaskCreatedMessage creates a new task created message
func NewTaskCreatedMessage(task *entities.Task) *QueueMessage {
	return &QueueMessage{
		ID:         task.ID().Value(),
		Type:       MessageTypeTaskCreated,
		Payload:    taskToPayload(task),
		Timestamp:  time.Now(),
		RetryCount: 0,
	}
}

// NewTaskUpdatedMessage creates a new task updated message
func NewTaskUpdatedMessage(task *entities.Task) *QueueMessage {
	return &QueueMessage{
		ID:         task.ID().Value(),
		Type:       MessageTypeTaskUpdated,
		Payload:    taskToPayload(task),
		Timestamp:  time.Now(),
		RetryCount: 0,
	}
}

// NewTaskResumedMessage creates a new task resumed message
func NewTaskResumedMessage(task *entities.Task) *QueueMessage {
	return &QueueMessage{
		ID:         task.ID().Value(),
		Type:       MessageTypeTaskResumed,
		Payload:    taskToPayload(task),
		Timestamp:  time.Now(),
		RetryCount: 0,
	}
}

// ToJSON serializes the message to JSON
func (m *QueueMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes a message from JSON
func FromJSON(data []byte) (*QueueMessage, error) {
	var message QueueMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// IncrementRetryCount increments the retry count for the message
func (m *QueueMessage) IncrementRetryCount() {
	m.RetryCount++
	m.Timestamp = time.Now()
}

// taskToPayload converts a task entity to a message payload
func taskToPayload(task *entities.Task) TaskPayload {
	return TaskPayload{
		TaskID:      task.ID().Value(),
		UserID:      task.UserID().Value(),
		Title:       task.Title(),
		Content:     task.Description(),
		Repository:  task.Repository().Value(),
		BranchName:  task.Branch().Value(),
		Status:      task.Status().Value(),
		Epic:        task.Epic(),
		Metadata:    task.Metadata(),
		CreatedAt:   task.CreatedAt(),
		UpdatedAt:   task.UpdatedAt(),
		StartedAt:   task.StartedAt(),
		CompletedAt: task.CompletedAt(),
		TokensUsed:  task.TokensUsed(),
		Version:     task.Version(),
	}
}

// IsValidMessageType checks if the message type is valid
func IsValidMessageType(msgType string) bool {
	switch MessageType(msgType) {
	case MessageTypeTaskCreated, MessageTypeTaskUpdated, MessageTypeTaskResumed:
		return true
	default:
		return false
	}
}

// GetMessageTypeDescription returns a human-readable description of the message type
func GetMessageTypeDescription(msgType MessageType) string {
	switch msgType {
	case MessageTypeTaskCreated:
		return "Task Created"
	case MessageTypeTaskUpdated:
		return "Task Updated"
	case MessageTypeTaskResumed:
		return "Task Resumed"
	default:
		return "Unknown"
	}
}