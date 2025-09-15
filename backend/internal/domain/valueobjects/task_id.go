package valueobjects

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// TaskID represents a unique identifier for a task
type TaskID struct {
	value string
}

var taskIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

// NewTaskID creates a new TaskID
func NewTaskID(value string) (TaskID, error) {
	if value == "" {
		return TaskID{}, fmt.Errorf("task ID cannot be empty")
	}

	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 100 {
		return TaskID{}, fmt.Errorf("task ID must be between 3 and 100 characters")
	}

	if !taskIDPattern.MatchString(value) {
		return TaskID{}, fmt.Errorf("task ID can only contain alphanumeric characters, hyphens, and underscores")
	}

	return TaskID{value: value}, nil
}

// GenerateTaskID creates a new unique TaskID using UUID
func GenerateTaskID() TaskID {
	id := uuid.New().String()
	return TaskID{value: "task-" + id}
}

// Value returns the string value of the TaskID
func (t TaskID) Value() string {
	return t.value
}

// String implements the Stringer interface
func (t TaskID) String() string {
	return t.value
}

// Equals checks if two TaskIDs are equal
func (t TaskID) Equals(other TaskID) bool {
	return t.value == other.value
}

// IsEmpty checks if the TaskID is empty
func (t TaskID) IsEmpty() bool {
	return t.value == ""
}