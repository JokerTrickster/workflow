package _interface

import "main/utils/db"

type ICreateTaskRepository interface {
	CreateTask(task *db.Task) error
}