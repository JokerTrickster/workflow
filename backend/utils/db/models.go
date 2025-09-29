package db

import (
	"time"
)

// Task represents a workflow task record
type Task struct {
	ID               uint64     `json:"id" gorm:"primaryKey;column:id"`
	RequestID        string     `json:"request_id" gorm:"column:request_id;type:varchar(255);uniqueIndex;not null"`
	Status           string     `json:"status" gorm:"column:status;index;not null;default:pending"`
	Tasks            string     `json:"tasks" gorm:"column:tasks;type:text;not null"`
	RepositoryName   string     `json:"repository_name" gorm:"column:repository_name;index;not null"`
	WorkingDir       *string    `json:"working_dir,omitempty" gorm:"column:working_dir"`
	Cmd              *string    `json:"cmd,omitempty" gorm:"column:cmd"`
	Provider         string     `json:"provider" gorm:"column:provider;index;not null"`
	Interactive      bool       `json:"interactive" gorm:"column:interactive;default:false"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at;index;not null"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updated_at;not null"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	ProcessingTimeMs *int64     `json:"processing_time_ms,omitempty" gorm:"column:processing_time_ms"`
	Result           *string    `json:"result,omitempty" gorm:"column:result;type:text"`
	Error            *string    `json:"error,omitempty" gorm:"column:error;type:text"`

	// GitHub Integration Fields
	GitHubIssueURL   *string    `json:"github_issue_url,omitempty" gorm:"column:github_issue_url"`
	GitHubPRURL      *string    `json:"github_pr_url,omitempty" gorm:"column:github_pr_url"`
	BranchName       *string    `json:"branch_name,omitempty" gorm:"column:branch_name"`
	CleanupStatus    *string    `json:"cleanup_status,omitempty" gorm:"column:cleanup_status"`
	ContinueTask     bool       `json:"continue_task" gorm:"column:continue_task;default:false"`
}

// TableName specifies the table name for Task model
func (Task) TableName() string {
	return "workflow_histories"
}

// TaskStatus constants for status field
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
	TaskStatusCancelled  = "cancelled"
)

// TaskProvider constants for provider field
const (
	TaskProviderClaude = "claude"
	TaskProviderCodex  = "codex"
	TaskProviderCursor = "cursor"
)