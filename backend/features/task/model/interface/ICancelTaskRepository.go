package _interface

import "main/utils/db"

type ICancelTaskRepository interface {
	GetTaskByRequestID(requestID string) (*db.Task, error)
	UpdateTaskStatus(requestID string, status string) error
}