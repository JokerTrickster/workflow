package _interface

import (
	"main/features/task/model/request"
	"main/features/task/model/response"
)

type IListTasksUseCase interface {
	ListTasks(req *request.ListTasksRequest) (*response.ListTasksResponse, error)
}