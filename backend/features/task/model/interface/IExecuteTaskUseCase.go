package _interface

import "main/features/task/model/response"

type IExecuteTaskUseCase interface {
	ExecuteTask(requestID string) (*response.ExecuteTaskResponse, error)
}