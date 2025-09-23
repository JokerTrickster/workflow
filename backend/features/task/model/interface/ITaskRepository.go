package _interface

import (
	"main/features/task/model/request"
	"main/utils/db"
)

// ITaskRepository defines the interface for task repository operations
type ITaskRepository interface {
	// CreateTask creates a new task record in the database
	CreateTask(task *db.Task) error

	// GetTaskByRequestID retrieves a task by its request ID
	GetTaskByRequestID(requestID string) (*db.Task, error)

	// GetTaskByID retrieves a task by its database ID
	GetTaskByID(id uint64) (*db.Task, error)

	// UpdateTaskStatus updates the status of a task
	UpdateTaskStatus(requestID string, status string) error

	// UpdateTaskToProcessing updates task status to processing and records start time
	UpdateTaskToProcessing(requestID string) error

	// UpdateTaskToCompleted updates task status to completed with result and timing
	UpdateTaskToCompleted(requestID string, result *string, processingTimeMs *int64) error

	// UpdateTaskToFailed updates task status to failed with error message
	UpdateTaskToFailed(requestID string, errorMsg string) error

	// CancelTask performs soft delete by updating status to cancelled
	CancelTask(requestID string) error

	// ListTasks retrieves tasks by repository with pagination (excluding cancelled)
	ListTasks(req *request.ListTasksRequest) ([]db.Task, int64, error)

	// GetActiveTasksCount returns count of non-cancelled tasks for a repository
	GetActiveTasksCount(repositoryName string) (int64, error)
}