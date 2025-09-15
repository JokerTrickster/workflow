package interfaces

import (
	"context"

	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/domain/valueobjects"
)

// TaskUsecase defines the interface for task use cases
type TaskUsecase interface {
	// Task creation and management
	CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.CreateTaskResponse, error)
	GetTask(ctx context.Context, taskID string, userID string) (*dto.GetTaskResponse, error)
	UpdateTask(ctx context.Context, taskID string, userID string, req dto.UpdateTaskRequest) (*dto.GetTaskResponse, error)
	DeleteTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error)
	
	// Task operations
	ListTasks(ctx context.Context, req dto.ListTasksRequest) (*dto.ListTasksResponse, error)
	ResumeTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error)
	CancelTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error)
	
	// Task monitoring and health
	GetTaskHealth(ctx context.Context, taskID string, userID string) (*dto.TaskHealthResponse, error)
	GetTaskStatistics(ctx context.Context, userID string) (*dto.TaskStatisticsResponse, error)
	GetQueueStatistics(ctx context.Context) (*dto.QueueStatisticsResponse, error)
	
	// Administrative operations
	GetTasksRequiringAttention(ctx context.Context) ([]dto.GetTaskResponse, error)
}

// TaskAuthorizationService defines the interface for task authorization
type TaskAuthorizationService interface {
	// CanAccessTask checks if the user can access the task
	CanAccessTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error
	
	// CanModifyTask checks if the user can modify the task
	CanModifyTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error
	
	// CanDeleteTask checks if the user can delete the task
	CanDeleteTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error
	
	// CanListTasks checks if the user can list tasks with the given filter
	CanListTasks(ctx context.Context, userID valueobjects.UserID, filter *dto.ListTasksRequest) error
	
	// CanViewStatistics checks if the user can view task statistics
	CanViewStatistics(ctx context.Context, userID valueobjects.UserID, targetUserID *valueobjects.UserID) error
}

// TaskEventService defines the interface for task event handling
type TaskEventService interface {
	// PublishTaskCreated publishes a task created event
	PublishTaskCreated(ctx context.Context, task *dto.GetTaskResponse) error
	
	// PublishTaskUpdated publishes a task updated event
	PublishTaskUpdated(ctx context.Context, task *dto.GetTaskResponse) error
	
	// PublishTaskDeleted publishes a task deleted event
	PublishTaskDeleted(ctx context.Context, taskID string, userID string) error
	
	// PublishTaskStatusChanged publishes a task status change event
	PublishTaskStatusChanged(ctx context.Context, taskID string, oldStatus string, newStatus string) error
	
	// PublishTaskResumed publishes a task resumed event
	PublishTaskResumed(ctx context.Context, taskID string, reason string) error
}