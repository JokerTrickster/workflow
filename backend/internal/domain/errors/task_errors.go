package errors

import (
	"fmt"
)

// TaskError represents a domain-specific task error
type TaskError struct {
	Code    string
	Message string
	Cause   error
}

func (e TaskError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e TaskError) Unwrap() error {
	return e.Cause
}

// Task error codes
const (
	ErrCodeTaskNotFound            = "TASK_NOT_FOUND"
	ErrCodeTaskAlreadyExists       = "TASK_ALREADY_EXISTS"
	ErrCodeInvalidTaskTitle        = "INVALID_TASK_TITLE"
	ErrCodeInvalidTaskDescription  = "INVALID_TASK_DESCRIPTION"
	ErrCodeInvalidStatusTransition = "INVALID_STATUS_TRANSITION"
	ErrCodeTaskValidationFailed    = "TASK_VALIDATION_FAILED"
	ErrCodeTaskAlreadyCompleted    = "TASK_ALREADY_COMPLETED"
	ErrCodeTaskNotActive           = "TASK_NOT_ACTIVE"
	ErrCodeConcurrentModification  = "CONCURRENT_MODIFICATION"
)

// NewTaskNotFoundError creates a task not found error
func NewTaskNotFoundError(taskID string) TaskError {
	return TaskError{
		Code:    ErrCodeTaskNotFound,
		Message: fmt.Sprintf("task with ID '%s' not found", taskID),
	}
}

// NewTaskAlreadyExistsError creates a task already exists error
func NewTaskAlreadyExistsError(taskID string) TaskError {
	return TaskError{
		Code:    ErrCodeTaskAlreadyExists,
		Message: fmt.Sprintf("task with ID '%s' already exists", taskID),
	}
}

// NewInvalidTaskTitleError creates an invalid task title error
func NewInvalidTaskTitleError(reason string) TaskError {
	return TaskError{
		Code:    ErrCodeInvalidTaskTitle,
		Message: fmt.Sprintf("invalid task title: %s", reason),
	}
}

// NewInvalidTaskDescriptionError creates an invalid task description error
func NewInvalidTaskDescriptionError(reason string) TaskError {
	return TaskError{
		Code:    ErrCodeInvalidTaskDescription,
		Message: fmt.Sprintf("invalid task description: %s", reason),
	}
}

// NewInvalidStatusTransitionError creates an invalid status transition error
func NewInvalidStatusTransitionError(fromStatus, toStatus string) TaskError {
	return TaskError{
		Code:    ErrCodeInvalidStatusTransition,
		Message: fmt.Sprintf("cannot transition from status '%s' to '%s'", fromStatus, toStatus),
	}
}

// NewTaskValidationFailedError creates a task validation error
func NewTaskValidationFailedError(reason string, cause error) TaskError {
	return TaskError{
		Code:    ErrCodeTaskValidationFailed,
		Message: fmt.Sprintf("task validation failed: %s", reason),
		Cause:   cause,
	}
}

// NewTaskAlreadyCompletedError creates a task already completed error
func NewTaskAlreadyCompletedError(taskID string) TaskError {
	return TaskError{
		Code:    ErrCodeTaskAlreadyCompleted,
		Message: fmt.Sprintf("task '%s' is already completed", taskID),
	}
}

// NewTaskNotActiveError creates a task not active error
func NewTaskNotActiveError(taskID, status string) TaskError {
	return TaskError{
		Code:    ErrCodeTaskNotActive,
		Message: fmt.Sprintf("task '%s' is not active (current status: %s)", taskID, status),
	}
}

// NewConcurrentModificationError creates a concurrent modification error
func NewConcurrentModificationError(taskID string) TaskError {
	return TaskError{
		Code:    ErrCodeConcurrentModification,
		Message: fmt.Sprintf("task '%s' was modified by another process", taskID),
	}
}

// IsTaskNotFound checks if the error is a task not found error
func IsTaskNotFound(err error) bool {
	if taskErr, ok := err.(TaskError); ok {
		return taskErr.Code == ErrCodeTaskNotFound
	}
	return false
}

// IsTaskAlreadyExists checks if the error is a task already exists error
func IsTaskAlreadyExists(err error) bool {
	if taskErr, ok := err.(TaskError); ok {
		return taskErr.Code == ErrCodeTaskAlreadyExists
	}
	return false
}

// IsInvalidStatusTransition checks if the error is an invalid status transition error
func IsInvalidStatusTransition(err error) bool {
	if taskErr, ok := err.(TaskError); ok {
		return taskErr.Code == ErrCodeInvalidStatusTransition
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if taskErr, ok := err.(TaskError); ok {
		return taskErr.Code == ErrCodeTaskValidationFailed ||
			taskErr.Code == ErrCodeInvalidTaskTitle ||
			taskErr.Code == ErrCodeInvalidTaskDescription
	}
	return false
}