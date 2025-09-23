package _interface

import "github.com/labstack/echo/v4"

// ICancelTaskHandler defines the interface for cancel task handler
type ICancelTaskHandler interface {
	CancelTask(c echo.Context) error
}