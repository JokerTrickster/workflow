package handler

import (
	"main/features/task/repository"
	"main/features/task/usecase"

	"github.com/labstack/echo/v4"
)

func NewTaskHandler(e *echo.Echo) {
	// Repository와 UseCase 초기화
	taskRepo := repository.NewTaskRepository()
	taskUC := usecase.NewTaskUseCase()

	// UseCase에 Repository 주입
	if uc, ok := taskUC.(*usecase.TaskUseCase); ok {
		uc.SetRepository(taskRepo)
	}

	// 각 API Handler 초기화
	NewCreateTaskHandler(e, taskUC)
	NewExecuteTaskHandler(e, taskUC)
	NewCancelTaskHandler(e, taskUC)
	NewListTasksHandler(e, taskUC)
	NewGetTaskStatusHandler(e, taskUC)
}