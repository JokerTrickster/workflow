package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/utils/db"

	"gorm.io/gorm"
)

type GetTaskStatusRepository struct {
	db *gorm.DB
}

func NewGetTaskStatusRepository() _interface.IGetTaskStatusRepository {
	return &GetTaskStatusRepository{
		db: db.GetDB(),
	}
}

func (r *GetTaskStatusRepository) GetTaskByRequestID(requestID string) (*db.Task, error) {
	var task db.Task
	if err := r.db.Where("request_id = ?", requestID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with request_id: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}