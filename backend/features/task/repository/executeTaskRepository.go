package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/utils/db"

	"gorm.io/gorm"
)

type ExecuteTaskRepository struct {
	db *gorm.DB
}

func NewExecuteTaskRepository() _interface.IExecuteTaskRepository {
	return &ExecuteTaskRepository{
		db: db.GetDB(),
	}
}

func (r *ExecuteTaskRepository) GetTaskByRequestID(requestID string) (*db.Task, error) {
	var task db.Task
	if err := r.db.Where("request_id = ?", requestID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with request_id: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (r *ExecuteTaskRepository) UpdateTaskStatus(requestID string, status string) error {
	if err := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}