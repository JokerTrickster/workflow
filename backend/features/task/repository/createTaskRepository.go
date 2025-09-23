package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/utils/db"

	"gorm.io/gorm"
)

type CreateTaskRepository struct {
	db *gorm.DB
}

func NewCreateTaskRepository() _interface.ICreateTaskRepository {
	return &CreateTaskRepository{
		db: db.GetDB(),
	}
}

func (r *CreateTaskRepository) CreateTask(task *db.Task) error {
	if err := r.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}