package _interface

import (
	"main/features/task/model/request"
	"main/features/task/model/response"
)

// ITaskUseCase defines the interface for task use case operations
type ITaskUseCase interface {
	// CreateTask creates a new task and returns the response
	CreateTask(req *request.CreateTaskRequest) (*response.CreateTaskResponse, error)

	// ExecuteTask executes a task by ID (queue to RabbitMQ and update status)
	ExecuteTask(requestID string) (*response.ExecuteTaskResponse, error)

	// CancelTask cancels a task by ID (soft delete)
	CancelTask(requestID string) (*response.CancelTaskResponse, error)

	// ListTasks retrieves tasks by repository with pagination
	ListTasks(req *request.ListTasksRequest) (*response.ListTasksResponse, error)

	// GetTaskStatus gets the current status of a task
	GetTaskStatus(requestID string) (*response.TaskResponse, error)
}