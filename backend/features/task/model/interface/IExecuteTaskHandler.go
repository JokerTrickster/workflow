package _interface

import "github.com/labstack/echo/v4"

// IExecuteTaskHandler defines the interface for execute task handler
type IExecuteTaskHandler interface {
	ExecuteTask(c echo.Context) error
}