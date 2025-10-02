package handler

import (
	"github.com/labstack/echo/v4"
)

func InitEpicTaskHandler(e *echo.Echo) *EpicTaskHandler {
	handler := NewEpicTaskHandler()

	// Epic tasks API route group
	epicTaskGroup := e.Group("/api/v1/epics/tasks")

	// Epic tasks endpoints
	epicTaskGroup.GET("", handler.GetAllTasks)
	epicTaskGroup.POST("", handler.CreateTask)
	epicTaskGroup.GET("/:id", handler.GetTask)
	epicTaskGroup.PUT("/:id", handler.UpdateTask)
	epicTaskGroup.DELETE("/:id", handler.DeleteTask)

	return handler
}
