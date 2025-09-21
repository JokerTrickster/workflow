package _interface

import "github.com/labstack/echo/v4"

type IRunTasksClaudeHandler interface {
	RunTasks(c echo.Context) error
}

type ICloneRepositoriesHandler interface {
	CloneRepositories(c echo.Context) error
}
