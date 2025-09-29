package request

// CreateTaskRequest represents the request structure for creating a new task
type CreateTaskRequest struct {
	Tasks          string `json:"tasks" validate:"required"`           // 실행할 작업 내용
	RepositoryName string `json:"repository_name" validate:"required"` // 레포지토리 이름 (필수)
	WorkingDir     string `json:"working_dir,omitempty"`               // 작업 디렉토리 (옵션)
	BranchName     string `json:"branch_name,omitempty"`               // 브랜치 이름 (옵션)
	Interactive    bool   `json:"interactive,omitempty"`               // 대화형 모드: 여러 작업을 순차 실행
	Cmd            string `json:"cmd,omitempty"`                       // 명령어 (옵션)
	Provider       string `json:"provider" validate:"required,oneof=claude codex cursor"` // AI 제공자 (필수)
}