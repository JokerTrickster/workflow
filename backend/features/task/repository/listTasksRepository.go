package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/request"
	"main/utils/db"

	"gorm.io/gorm"
)

type ListTasksRepository struct {
	db *gorm.DB
}

func NewListTasksRepository() _interface.IListTasksRepository {
	return &ListTasksRepository{
		db: db.GetDB(),
	}
}

func (r *ListTasksRepository) GetTasks(params *request.ListTasksRequest) ([]*db.Task, int64, error) {
	var tasks []*db.Task
	var totalCount int64

	query := r.db.Model(&db.Task{}).Where("repository_name = ? AND status != ?", params.RepositoryName, "cancelled")

	// Count total records
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Apply pagination and ordering
	offset := (params.Page - 1) * params.Limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.Limit).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
	}

	return tasks, totalCount, nil
}