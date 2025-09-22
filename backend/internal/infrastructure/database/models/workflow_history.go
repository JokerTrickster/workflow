package models

import (
	"time"
)

// WorkflowHistory represents a workflow task history record
type WorkflowHistory struct {
	ID               uint64     `json:"id" gorm:"primaryKey;column:id"`
	RequestID        string     `json:"request_id" gorm:"column:request_id;uniqueIndex;not null"`
	Status           string     `json:"status" gorm:"column:status;index;not null;default:pending"`
	Tasks            string     `json:"tasks" gorm:"column:tasks;type:text;not null"`
	RepositoryName   string     `json:"repository_name" gorm:"column:repository_name;index;not null"`
	WorkingDir       *string    `json:"working_dir,omitempty" gorm:"column:working_dir"`
	ClaudeCmd        *string    `json:"claude_cmd,omitempty" gorm:"column:claude_cmd"`
	Interactive      bool       `json:"interactive" gorm:"column:interactive;default:false"`
	ContinueTask     bool       `json:"continue_task" gorm:"column:continue_task;default:false"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at;index;not null"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	ProcessingTimeMs *int64     `json:"processing_time_ms,omitempty" gorm:"column:processing_time_ms"`
	Result           *string    `json:"result,omitempty" gorm:"column:result;type:text"`
	Error            *string    `json:"error,omitempty" gorm:"column:error;type:text"`
}

// TableName specifies the table name for WorkflowHistory model
func (WorkflowHistory) TableName() string {
	return "workflow_histories"
}

// WorkflowStatus constants for status field
const (
	WorkflowStatusPending    = "pending"
	WorkflowStatusProcessing = "processing"
	WorkflowStatusCompleted  = "completed"
	WorkflowStatusFailed     = "failed"
	WorkflowStatusCancelled  = "cancelled"
)