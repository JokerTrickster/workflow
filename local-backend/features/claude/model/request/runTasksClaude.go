package request

type ReqRunTasksClaude struct {
	Tasks          string `json:"tasks" validate:"required"`              // 실행할 작업 내용
	RepositoryName string `json:"repository_name" validate:"required"`    // 레포지토리 이름 (필수)
	WorkingDir     string `json:"working_dir,omitempty"`                  // 작업 디렉토리 (옵션)
	Interactive    bool   `json:"interactive,omitempty"`                  // 대화형 모드: 여러 작업을 순차 실행
	ClaudeCmd      string `json:"claude_cmd,omitempty"`                   // Claude CLI 명령어 경로 (옵션)
	ContinueTask   bool   `json:"continue_task,omitempty"`                // 기존 작업 이어서 하기 (옵션)
}
