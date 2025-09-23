package _interface

import "main/utils/db"

type IExecuteTaskRepository interface {
	GetTaskByRequestID(requestID string) (*db.Task, error)
	UpdateTaskStatus(requestID string, status string) error
}