package _interface

import "github.com/labstack/echo/v4"

// ITaskHandler defines the interface for task handler operations
type ITaskHandler interface {
	// CreateTask handles POST /api/v1/tasks
	CreateTask(c echo.Context) error

	// ExecuteTask handles POST /api/v1/tasks/{id}/execute
	ExecuteTask(c echo.Context) error

	// CancelTask handles DELETE /api/v1/tasks/{id}
	CancelTask(c echo.Context) error

	// ListTasks handles GET /api/v1/tasks
	ListTasks(c echo.Context) error

	// GetTaskStatus handles GET /api/v1/tasks/{id}/status
	GetTaskStatus(c echo.Context) error
}