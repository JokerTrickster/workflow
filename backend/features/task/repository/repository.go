package repository

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/request"
	"main/utils/db"
	"time"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository() _interface.ITaskRepository {
	return &TaskRepository{
		db: db.GetDB(),
	}
}

func (r *TaskRepository) CreateTask(task *db.Task) error {
	if err := r.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetTaskByRequestID(requestID string) (*db.Task, error) {
	var task db.Task
	if err := r.db.Where("request_id = ?", requestID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with request_id: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) GetTaskByID(id uint64) (*db.Task, error) {
	var task db.Task
	if err := r.db.Where("id = ?", id).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) UpdateTaskStatus(requestID string, status string) error {
	result := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update task status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no task found with request_id: %s", requestID)
	}
	return nil
}

func (r *TaskRepository) UpdateTaskToProcessing(requestID string) error {
	updates := map[string]interface{}{
		"status":     db.TaskStatusProcessing,
		"updated_at": time.Now(),
	}

	result := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update task to processing: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no task found with request_id: %s", requestID)
	}
	return nil
}

func (r *TaskRepository) UpdateTaskToCompleted(requestID string, result *string, processingTimeMs *int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":             db.TaskStatusCompleted,
		"updated_at":         now,
		"completed_at":       &now,
		"result":             result,
		"processing_time_ms": processingTimeMs,
	}

	dbResult := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Updates(updates)
	if dbResult.Error != nil {
		return fmt.Errorf("failed to update task to completed: %w", dbResult.Error)
	}
	if dbResult.RowsAffected == 0 {
		return fmt.Errorf("no task found with request_id: %s", requestID)
	}
	return nil
}

func (r *TaskRepository) UpdateTaskToFailed(requestID string, errorMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       db.TaskStatusFailed,
		"updated_at":   now,
		"completed_at": &now,
		"error":        errorMsg,
	}

	result := r.db.Model(&db.Task{}).Where("request_id = ?", requestID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update task to failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no task found with request_id: %s", requestID)
	}
	return nil
}

func (r *TaskRepository) CancelTask(requestID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       db.TaskStatusCancelled,
		"updated_at":   now,
		"completed_at": &now,
	}

	result := r.db.Model(&db.Task{}).Where("request_id = ? AND status != ?", requestID, db.TaskStatusCancelled).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to cancel task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no task found with request_id: %s or task already cancelled", requestID)
	}
	return nil
}

func (r *TaskRepository) ListTasks(req *request.ListTasksRequest) ([]db.Task, int64, error) {
	var tasks []db.Task
	var total int64

	// 기본값 설정
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 기본 쿼리 (취소된 태스크 제외, repository 필터링)
	baseQuery := r.db.Model(&db.Task{}).Where("repository_name = ? AND status != ?", req.RepositoryName, db.TaskStatusCancelled)

	// 총 개수 조회
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// 페이지네이션과 정렬 적용
	offset := (page - 1) * limit
	if err := baseQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

func (r *TaskRepository) GetActiveTasksCount(repositoryName string) (int64, error) {
	var count int64
	if err := r.db.Model(&db.Task{}).Where("repository_name = ? AND status != ?", repositoryName, db.TaskStatusCancelled).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active tasks: %w", err)
	}
	return count, nil
}