package _interface

import (
	"main/features/task/model/request"
	"main/utils/db"
)

type IListTasksRepository interface {
	GetTasks(params *request.ListTasksRequest) ([]*db.Task, int64, error)
}