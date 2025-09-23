package _interface

import "github.com/labstack/echo/v4"

// IGetTaskStatusHandler defines the interface for get task status handler
type IGetTaskStatusHandler interface {
	GetTaskStatus(c echo.Context) error
}