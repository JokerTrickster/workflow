package response

import "time"

// TaskResponse represents the response structure for task operations
type TaskResponse struct {
	ID               uint64     `json:"id"`
	RequestID        string     `json:"request_id"`
	Status           string     `json:"status"`
	Tasks            string     `json:"tasks"`
	RepositoryName   string     `json:"repository_name"`
	WorkingDir       *string    `json:"working_dir,omitempty"`
	Cmd              *string    `json:"cmd,omitempty"`
	Provider         string     `json:"provider"`
	Interactive      bool       `json:"interactive"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ProcessingTimeMs *int64     `json:"processing_time_ms,omitempty"`
	Result           *string    `json:"result,omitempty"`
	Error            *string    `json:"error,omitempty"`
}

// CreateTaskResponse represents the response structure for task creation
type CreateTaskResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// ExecuteTaskResponse represents the response structure for task execution
type ExecuteTaskResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
}

// CancelTaskResponse represents the response structure for task cancellation
type CancelTaskResponse struct {
	RequestID   string `json:"request_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	CancelledAt string `json:"cancelled_at"`
}

// ListTasksResponse represents the response structure for task listing
type ListTasksResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	HasMore    bool           `json:"has_more"`
}