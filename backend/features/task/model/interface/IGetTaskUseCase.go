package _interface

import "main/features/task/model/response"

type IGetTaskUseCase interface {
	GetTask(requestID string) (*response.TaskResponse, error)
}