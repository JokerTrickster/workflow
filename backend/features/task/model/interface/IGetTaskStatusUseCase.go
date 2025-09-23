package _interface

import "main/features/task/model/response"

type IGetTaskStatusUseCase interface {
	GetTaskStatus(requestID string) (*response.TaskResponse, error)
}