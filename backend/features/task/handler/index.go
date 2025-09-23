package handler

import (
	"main/features/task/repository"
	"main/features/task/usecase"

	"github.com/labstack/echo/v4"
)

func NewTaskHandler(e *echo.Echo) {
	// 개별 Repository 초기화
	createTaskRepo := repository.NewCreateTaskRepository()
	executeTaskRepo := repository.NewExecuteTaskRepository()
	cancelTaskRepo := repository.NewCancelTaskRepository()
	listTasksRepo := repository.NewListTasksRepository()
	getTaskStatusRepo := repository.NewGetTaskStatusRepository()

	// 개별 UseCase 초기화
	createTaskUC := usecase.NewCreateTaskUseCase(createTaskRepo)
	executeTaskUC := usecase.NewExecuteTaskUseCase(executeTaskRepo)
	cancelTaskUC := usecase.NewCancelTaskUseCase(cancelTaskRepo)
	listTasksUC := usecase.NewListTasksUseCase(listTasksRepo)
	getTaskStatusUC := usecase.NewGetTaskStatusUseCase(getTaskStatusRepo)

	// 각 API Handler 초기화
	NewCreateTaskHandler(e, createTaskUC)
	NewExecuteTaskHandler(e, executeTaskUC)
	NewCancelTaskHandler(e, cancelTaskUC)
	NewListTasksHandler(e, listTasksUC)
	NewGetTaskStatusHandler(e, getTaskStatusUC)
}
