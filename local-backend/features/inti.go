package features

import (
	"net/http"

	claudeHandler "main/features/claude/handler"

	"github.com/labstack/echo/v4"
)

func InitHandler(e *echo.Echo) error {
	//elb 헬스체크용
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Tasks API - 간단한 더미 응답
	e.GET("/api/v1/tasks", func(c echo.Context) error {
		repositoryName := c.QueryParam("repository_name")
		if repositoryName == "" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "repository_name query parameter is required",
			})
		}

		// 빈 tasks 배열 반환 (나중에 실제 데이터로 교체 가능)
		return c.JSON(http.StatusOK, map[string]any{
			"tasks": []any{},
			"total": 0,
			"page":  1,
			"limit": 20,
		})
	})

	//인증 핸들러 초기화
	claudeHandler.NewClaudeHandler(e)

	return nil
}
