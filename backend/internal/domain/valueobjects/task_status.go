package valueobjects

import (
	"fmt"
	"strings"
)

// TaskStatus represents the status of a task in its lifecycle
type TaskStatus struct {
	value string
}

// Valid task statuses
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
)

var validStatuses = map[string]bool{
	StatusPending:    true,
	StatusProcessing: true,
	StatusCompleted:  true,
	StatusFailed:     true,
	StatusCancelled:  true,
}

// Valid status transitions
var validTransitions = map[string][]string{
	StatusPending: {StatusProcessing, StatusCancelled},
	StatusProcessing: {StatusCompleted, StatusFailed, StatusCancelled},
	StatusCompleted: {}, // Terminal state
	StatusFailed: {StatusPending}, // Can be retried
	StatusCancelled: {StatusPending}, // Can be restarted
}

// NewTaskStatus creates a new TaskStatus
func NewTaskStatus(value string) (TaskStatus, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	if !validStatuses[value] {
		return TaskStatus{}, fmt.Errorf("invalid task status: %s. Valid statuses are: pending, processing, completed, failed, cancelled", value)
	}

	return TaskStatus{value: value}, nil
}

// NewPendingStatus creates a new pending status (default for new tasks)
func NewPendingStatus() TaskStatus {
	return TaskStatus{value: StatusPending}
}

// Value returns the string value of the TaskStatus
func (s TaskStatus) Value() string {
	return s.value
}

// String implements the Stringer interface
func (s TaskStatus) String() string {
	return s.value
}

// Equals checks if two TaskStatuses are equal
func (s TaskStatus) Equals(other TaskStatus) bool {
	return s.value == other.value
}

// CanTransitionTo checks if the current status can transition to the target status
func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	allowedTransitions := validTransitions[s.value]
	for _, allowed := range allowedTransitions {
		if allowed == target.value {
			return true
		}
	}
	return false
}

// IsPending checks if the status is pending
func (s TaskStatus) IsPending() bool {
	return s.value == StatusPending
}

// IsProcessing checks if the status is processing
func (s TaskStatus) IsProcessing() bool {
	return s.value == StatusProcessing
}

// IsCompleted checks if the status is completed
func (s TaskStatus) IsCompleted() bool {
	return s.value == StatusCompleted
}

// IsFailed checks if the status is failed
func (s TaskStatus) IsFailed() bool {
	return s.value == StatusFailed
}

// IsCancelled checks if the status is cancelled
func (s TaskStatus) IsCancelled() bool {
	return s.value == StatusCancelled
}

// IsTerminal checks if the status is a terminal state (completed)
func (s TaskStatus) IsTerminal() bool {
	return s.value == StatusCompleted
}

// IsActive checks if the status represents an active task (pending or processing)
func (s TaskStatus) IsActive() bool {
	return s.value == StatusPending || s.value == StatusProcessing
}

// GetValidTransitions returns all valid transitions from the current status
func (s TaskStatus) GetValidTransitions() []TaskStatus {
	allowedTransitions := validTransitions[s.value]
	result := make([]TaskStatus, len(allowedTransitions))
	for i, transition := range allowedTransitions {
		result[i] = TaskStatus{value: transition}
	}
	return result
}