package handler

import (
	"github.com/labstack/echo/v4"
)

func InitWorkLogHandler(e *echo.Echo) *WorkLogHandler {
	handler := NewWorkLogHandler()

	// Work logs API route group
	workLogGroup := e.Group("/api/v1/work-logs")

	// Work logs endpoints
	workLogGroup.GET("", handler.GetWorkLogs)
	workLogGroup.POST("/entry", handler.CreateWorkLogEntry)

	return handler
}
