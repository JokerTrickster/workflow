package _interface

import "github.com/labstack/echo/v4"

// ICreateTaskHandler defines the interface for create task handler
type ICreateTaskHandler interface {
	CreateTask(c echo.Context) error
}