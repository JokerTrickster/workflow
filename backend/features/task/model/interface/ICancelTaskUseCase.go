package _interface

import "main/features/task/model/response"

type ICancelTaskUseCase interface {
	CancelTask(requestID string) (*response.CancelTaskResponse, error)
}