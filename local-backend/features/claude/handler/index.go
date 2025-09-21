package handler

import (
	"main/features/claude/repository"
	"main/features/claude/usecase"
	"main/utils/db/mysql"
	"time"

	"github.com/labstack/echo/v4"
)

func NewClaudeHandler(c *echo.Echo) {
	NewRunTasksClaudeHandler(c, usecase.NewRunTasksClaudeUseCase(repository.NewRunTasksClaudeRepository(mysql.GormMysqlDB), mysql.DBTimeOut*time.Second))
	NewCloneRepositoriesHandler(c, usecase.NewCloneRepositoriesUseCase(repository.NewCloneRepositoriesRepository(mysql.GormMysqlDB), mysql.DBTimeOut*time.Second))
}
