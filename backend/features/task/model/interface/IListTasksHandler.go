package _interface

import "github.com/labstack/echo/v4"

// IListTasksHandler defines the interface for list tasks handler
type IListTasksHandler interface {
	ListTasks(c echo.Context) error
}