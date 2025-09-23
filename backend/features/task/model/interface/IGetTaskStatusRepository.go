package _interface

import "main/utils/db"

type IGetTaskStatusRepository interface {
	GetTaskByRequestID(requestID string) (*db.Task, error)
}