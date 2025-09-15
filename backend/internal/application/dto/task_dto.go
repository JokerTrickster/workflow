package dto

import (
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
)

// CreateTaskRequest represents the request to create a new task
type CreateTaskRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	Title       string `json:"title" validate:"required,min=1,max=500"`
	Description string `json:"description" validate:"max=5000"`
	Repository  string `json:"repository" validate:"required"`
	Epic        string `json:"epic" validate:"required,min=1,max=255"`
	Branch      string `json:"branch" validate:"required"`
}

// CreateTaskResponse represents the response after creating a task
type CreateTaskResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetTaskResponse represents the response for getting a task
type GetTaskResponse struct {
	TaskID      string            `json:"task_id"`
	UserID      string            `json:"user_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Repository  string            `json:"repository"`
	Epic        string            `json:"epic"`
	Branch      string            `json:"branch"`
	TokensUsed  int               `json:"tokens_used"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	StartedAt   *string           `json:"started_at,omitempty"`
	CompletedAt *string           `json:"completed_at,omitempty"`
	Metadata    map[string]string `json:"metadata"`
	Version     int64             `json:"version"`
}

// UpdateTaskRequest represents the request to update a task
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`
	Version     int64   `json:"version" validate:"required"`
}

// ListTasksRequest represents the request to list tasks with filtering
type ListTasksRequest struct {
	UserID           *string    `json:"user_id,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Repository       *string    `json:"repository,omitempty"`
	Epic             *string    `json:"epic,omitempty"`
	Branch           *string    `json:"branch,omitempty"`
	CreatedAfter     *time.Time `json:"created_after,omitempty"`
	CreatedBefore    *time.Time `json:"created_before,omitempty"`
	TitleContains    *string    `json:"title_contains,omitempty"`
	OrderBy          string     `json:"order_by" validate:"omitempty,oneof=created_at updated_at title status tokens_used"`
	OrderDirection   string     `json:"order_direction" validate:"omitempty,oneof=ASC DESC"`
	Limit            int        `json:"limit" validate:"min=1,max=100"`
	Offset           int        `json:"offset" validate:"min=0"`
}

// ListTasksResponse represents the response for listing tasks
type ListTasksResponse struct {
	Tasks      []GetTaskResponse `json:"tasks"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	HasMore    bool              `json:"has_more"`
}

// TaskHealthResponse represents the health status of a task
type TaskHealthResponse struct {
	TaskID  string   `json:"task_id"`
	Health  string   `json:"health"`
	Message string   `json:"message"`
	Issues  []string `json:"issues"`
}

// TaskStatisticsResponse represents aggregated statistics for a user's tasks
type TaskStatisticsResponse struct {
	UserID               string  `json:"user_id"`
	TotalTasks           int     `json:"total_tasks"`
	CompletedTasks       int     `json:"completed_tasks"`
	FailedTasks          int     `json:"failed_tasks"`
	PendingTasks         int     `json:"pending_tasks"`
	ProcessingTasks      int     `json:"processing_tasks"`
	CancelledTasks       int     `json:"cancelled_tasks"`
	TotalTokensUsed      int     `json:"total_tokens_used"`
	AverageTokensPerTask int     `json:"average_tokens_per_task"`
	CompletionRate       float64 `json:"completion_rate"`
	LastActivityAt       *string `json:"last_activity_at,omitempty"`
}

// QueueStatisticsResponse represents queue performance metrics
type QueueStatisticsResponse struct {
	TotalEnqueued         int64   `json:"total_enqueued"`
	TotalDequeued         int64   `json:"total_dequeued"`
	TotalProcessed        int64   `json:"total_processed"`
	TotalFailed           int64   `json:"total_failed"`
	CurrentQueueLength    int     `json:"current_queue_length"`
	AverageProcessingTime string  `json:"average_processing_time"`
	ThroughputPerMinute   float64 `json:"throughput_per_minute"`
	ErrorRate             float64 `json:"error_rate"`
	DeadLetterCount       int     `json:"dead_letter_count"`
	WorkersActive         int     `json:"workers_active"`
	LastActivityAt        string  `json:"last_activity_at"`
}

// TaskActionRequest represents requests for task actions (cancel, resume)
type TaskActionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// TaskActionResponse represents responses for task actions
type TaskActionResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Helper functions to convert between domain entities and DTOs

// ToGetTaskResponse converts a domain task entity to a DTO response
func ToGetTaskResponse(task *entities.Task) GetTaskResponse {
	var startedAt, completedAt *string
	if task.StartedAt() != nil {
		s := task.StartedAt().Format(time.RFC3339)
		startedAt = &s
	}
	if task.CompletedAt() != nil {
		s := task.CompletedAt().Format(time.RFC3339)
		completedAt = &s
	}

	return GetTaskResponse{
		TaskID:      task.ID().Value(),
		UserID:      task.UserID().Value(),
		Title:       task.Title(),
		Description: task.Description(),
		Status:      task.Status().Value(),
		Repository:  task.Repository().Value(),
		Epic:        task.Epic(),
		Branch:      task.Branch().Value(),
		TokensUsed:  task.TokensUsed(),
		CreatedAt:   task.CreatedAt().Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt().Format(time.RFC3339),
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Metadata:    task.Metadata(),
		Version:     task.Version(),
	}
}

// ToCreateTaskResponse converts task creation result to response
func ToCreateTaskResponse(task *entities.Task) CreateTaskResponse {
	return CreateTaskResponse{
		TaskID:    task.ID().Value(),
		Status:    task.Status().Value(),
		CreatedAt: task.CreatedAt().Format(time.RFC3339),
	}
}

// ToListTasksResponse converts a slice of tasks to list response
func ToListTasksResponse(tasks []*entities.Task, total, limit, offset int) ListTasksResponse {
	taskResponses := make([]GetTaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = ToGetTaskResponse(task)
	}

	hasMore := offset+len(tasks) < total

	return ListTasksResponse{
		Tasks:   taskResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}

// Validation helpers

// Validate validates the create task request
func (r *CreateTaskRequest) Validate() error {
	if r.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "is required"}
	}
	if r.Title == "" {
		return &ValidationError{Field: "title", Message: "is required"}
	}
	if len(r.Title) > 500 {
		return &ValidationError{Field: "title", Message: "cannot exceed 500 characters"}
	}
	if len(r.Description) > 5000 {
		return &ValidationError{Field: "description", Message: "cannot exceed 5000 characters"}
	}
	if r.Repository == "" {
		return &ValidationError{Field: "repository", Message: "is required"}
	}
	if r.Epic == "" {
		return &ValidationError{Field: "epic", Message: "is required"}
	}
	if len(r.Epic) > 255 {
		return &ValidationError{Field: "epic", Message: "cannot exceed 255 characters"}
	}
	if r.Branch == "" {
		return &ValidationError{Field: "branch", Message: "is required"}
	}
	return nil
}

// ToValueObjects converts DTO request to domain value objects
func (r *CreateTaskRequest) ToValueObjects() (valueobjects.UserID, valueobjects.RepositoryPath, valueobjects.BranchName, error) {
	userID, err := valueobjects.NewUserID(r.UserID)
	if err != nil {
		return valueobjects.UserID{}, valueobjects.RepositoryPath{}, valueobjects.BranchName{}, err
	}

	repository, err := valueobjects.NewRepositoryPath(r.Repository)
	if err != nil {
		return valueobjects.UserID{}, valueobjects.RepositoryPath{}, valueobjects.BranchName{}, err
	}

	branch, err := valueobjects.NewBranchName(r.Branch)
	if err != nil {
		return valueobjects.UserID{}, valueobjects.RepositoryPath{}, valueobjects.BranchName{}, err
	}

	return userID, repository, branch, nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}