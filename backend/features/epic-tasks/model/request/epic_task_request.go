package request

type GetEpicTasksRequest struct {
	Repository string `query:"repository"`
}

type CreateEpicTaskRequest struct {
	Repository string       `json:"repository" validate:"required"`
	Task       EpicTaskFile `json:"task" validate:"required"`
}

type UpdateEpicTaskRequest struct {
	Repository string       `json:"repository" validate:"required"`
	Task       EpicTaskFile `json:"task" validate:"required"`
}

type EpicTaskFile struct {
	Metadata TaskFileMetadata `json:"metadata" validate:"required"`
	Content  string           `json:"content"`
}

type TaskFileMetadata struct {
	ID           string   `json:"id" yaml:"id" validate:"required"`
	Title        string   `json:"title" yaml:"title" validate:"required"`
	Status       string   `json:"status" yaml:"status"`
	Repository   string   `json:"repository" yaml:"repository"`
	Epic         string   `json:"epic" yaml:"epic"`
	Branch       string   `json:"branch,omitempty" yaml:"branch,omitempty"`
	CreatedAt    string   `json:"createdAt" yaml:"createdAt"`
	UpdatedAt    string   `json:"updatedAt" yaml:"updatedAt"`
	StartedAt    string   `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	CompletedAt  string   `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
	TokensUsed   int      `json:"tokensUsed" yaml:"tokensUsed"`
	GithubIssue  int      `json:"githubIssue,omitempty" yaml:"githubIssue,omitempty"`
	PrUrl        string   `json:"prUrl,omitempty" yaml:"prUrl,omitempty"`
	BuildStatus  string   `json:"buildStatus,omitempty" yaml:"buildStatus,omitempty"`
	LintStatus   string   `json:"lintStatus,omitempty" yaml:"lintStatus,omitempty"`
}
