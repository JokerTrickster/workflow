package request

type ReqCloneRepositories struct {
	GitHubUsername  string `json:"github_username,omitempty"`              // GitHub 사용자명/조직명 (기본값: JokerTrickster)
	TargetDirectory string `json:"target_directory,omitempty"`             // 대상 디렉토리 (옵션)
	GitHubToken     string `json:"github_token,omitempty"`                 // GitHub 토큰 (옵션)
}

type ResCloneRepositories struct {
	Status            string             `json:"status"`              // 전체 작업 상태 (success/failed/partial)
	TotalRepositories int                `json:"total_repositories"`  // 전체 저장소 수
	ClonedCount       int                `json:"cloned_count"`        // 복제된 저장소 수
	SkippedCount      int                `json:"skipped_count"`       // 건너뛴 저장소 수
	FailedCount       int                `json:"failed_count"`        // 실패한 저장소 수
	Details           []RepositoryResult `json:"details"`             // 개별 저장소 결과 상세
}

type RepositoryResult struct {
	Name         string `json:"name"`          // 저장소 이름
	CloneURL     string `json:"clone_url"`     // 복제 URL
	Status       string `json:"status"`        // 상태 (cloned/skipped/failed)
	ErrorMessage string `json:"error_message,omitempty"` // 에러 메시지 (실패시)
	LocalPath    string `json:"local_path,omitempty"`    // 로컬 경로 (성공시)
}