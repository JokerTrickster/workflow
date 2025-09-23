package _interface

import (
	"main/features/task/model/request"
	"main/features/task/model/response"
)

type ICreateTaskUseCase interface {
	CreateTask(req *request.CreateTaskRequest) (*response.CreateTaskResponse, error)
}