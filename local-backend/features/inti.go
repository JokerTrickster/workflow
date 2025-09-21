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
	//인증 핸들러 초기화
	claudeHandler.NewClaudeHandler(e)

	return nil
}
