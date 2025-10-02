package request

import "time"

type GetWorkLogsRequest struct {
	Repository string `query:"repository" validate:"required"`
	StartDate  string `query:"startDate"`
	EndDate    string `query:"endDate"`
}

type CreateWorkLogEntryRequest struct {
	Repository string       `json:"repository" validate:"required"`
	Entry      WorkLogEntry `json:"entry" validate:"required"`
}

type WorkLogEntry struct {
	TaskID            string            `json:"taskId" validate:"required"`
	TaskTitle         string            `json:"taskTitle" validate:"required"`
	Repository        string            `json:"repository" validate:"required"`
	Status            string            `json:"status" validate:"required,oneof=pending in_progress completed failed"`
	Timestamp         time.Time         `json:"timestamp"`
	ProgressUpdate    string            `json:"progressUpdate,omitempty"`
	IssuesDiscovered  []string          `json:"issuesDiscovered,omitempty"`
	ImprovementsMade  []string          `json:"improvementsMade,omitempty"`
	Metadata          *WorkLogMetadata  `json:"metadata,omitempty"`
}

type WorkLogMetadata struct {
	Branch       string `json:"branch,omitempty"`
	GithubIssue  int    `json:"githubIssue,omitempty"`
	PrUrl        string `json:"prUrl,omitempty"`
	TokensUsed   int    `json:"tokensUsed,omitempty"`
}
