package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/utils/db"

	"gorm.io/gorm"
)

type CancelTaskRepository struct {
	db *gorm.DB
}

func NewCancelTaskRepository() _interface.ICancelTaskRepository {
	return &CancelTaskRepository{
		db: db.GetDB(),
	}
}

func (r *CancelTaskRepository) GetTaskByRequestID(requestID string) (*db.Task, error) {
	var task db.Task
	if err := r.db.Where("request_id = ?", requestID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with request_id: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (r *CancelTaskRepository) UpdateTaskStatus(requestID string, status string) error {
	if err := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}