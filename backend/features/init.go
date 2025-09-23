package features

import (
	"net/http"

	taskHandler "main/features/task/handler"

	"github.com/labstack/echo/v4"
)

func InitHandler(e *echo.Echo) error {
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"service": "workflow-backend",
		})
	})

	// Task handler 초기화
	taskHandler.NewTaskHandler(e)

	return nil
}