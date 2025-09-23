package request

// ListTasksRequest represents the request structure for listing tasks
type ListTasksRequest struct {
	RepositoryName string `query:"repository_name" validate:"required"` // 레포지토리 이름으로 필터링 (필수)
	Page           int    `query:"page" validate:"min=1"`               // 페이지 번호 (기본값: 1)
	Limit          int    `query:"limit" validate:"min=1,max=100"`      // 페이지당 항목 수 (기본값: 20, 최대: 100)
}