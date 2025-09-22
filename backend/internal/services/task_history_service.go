package services

import (
	"context"
	"math"

	"gorm.io/gorm"

	"local-backend-server/internal/errors"
	"local-backend-server/internal/infrastructure/database/models"
)

// TaskHistoryService handles task history operations
type TaskHistoryService struct {
	db *gorm.DB
}

// NewTaskHistoryService creates a new TaskHistoryService
func NewTaskHistoryService(db *gorm.DB) *TaskHistoryService {
	return &TaskHistoryService{
		db: db,
	}
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// TaskHistoryResponse represents the paginated task history response
type TaskHistoryResponse struct {
	Data       []models.WorkflowHistory `json:"data"`
	Pagination PaginationMeta           `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ValidatePaginationParams validates and normalizes pagination parameters
func ValidatePaginationParams(page, limit int) (*PaginationParams, *errors.AppError) {
	// Validate page
	if page < 1 {
		return nil, errors.NewValidationError("page must be greater than 0").
			WithDetails("page parameter must be a positive integer")
	}

	// Validate limit
	if limit < 1 {
		return nil, errors.NewValidationError("limit must be greater than 0").
			WithDetails("limit parameter must be a positive integer")
	}

	if limit > 100 {
		return nil, errors.NewValidationError("limit exceeds maximum allowed value").
			WithDetails("limit parameter cannot exceed 100")
	}

	return &PaginationParams{
		Page:  page,
		Limit: limit,
	}, nil
}

// ValidateRepositoryName validates repository name format
func ValidateRepositoryName(repositoryName string) *errors.AppError {
	if repositoryName == "" {
		return errors.NewValidationError("repository name is required").
			WithDetails("repository_name parameter cannot be empty")
	}

	// Basic repository name validation (can be extended as needed)
	if len(repositoryName) > 255 {
		return errors.NewValidationError("repository name too long").
			WithDetails("repository_name parameter cannot exceed 255 characters")
	}

	return nil
}

// GetTaskHistory retrieves paginated task history for a repository
func (s *TaskHistoryService) GetTaskHistory(ctx context.Context, repositoryName string, params *PaginationParams) (*TaskHistoryResponse, *errors.AppError) {
	// Validate repository name
	if err := ValidateRepositoryName(repositoryName); err != nil {
		return nil, err
	}

	// Get total count for pagination metadata
	var total int64
	if err := s.db.WithContext(ctx).
		Model(&models.WorkflowHistory{}).
		Where("repository_name = ?", repositoryName).
		Count(&total).Error; err != nil {
		return nil, errors.NewDatabaseQueryError(err).
			WithDetails("failed to count task history records")
	}

	// If no records found, return appropriate response
	if total == 0 {
		return &TaskHistoryResponse{
			Data: []models.WorkflowHistory{},
			Pagination: PaginationMeta{
				Page:       params.Page,
				Limit:      params.Limit,
				Total:      0,
				TotalPages: 0,
			},
		}, nil
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	// Validate page doesn't exceed total pages
	if params.Page > totalPages {
		return nil, errors.NewValidationError("page exceeds available pages").
			WithDetails("requested page number exceeds total available pages")
	}

	// Query task history with pagination
	var tasks []models.WorkflowHistory
	offset := (params.Page - 1) * params.Limit

	if err := s.db.WithContext(ctx).
		Where("repository_name = ?", repositoryName).
		Order("created_at DESC").
		Offset(offset).
		Limit(params.Limit).
		Find(&tasks).Error; err != nil {
		return nil, errors.NewDatabaseQueryError(err).
			WithDetails("failed to retrieve task history records")
	}

	return &TaskHistoryResponse{
		Data: tasks,
		Pagination: PaginationMeta{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// CheckRepositoryExists checks if any tasks exist for the given repository
func (s *TaskHistoryService) CheckRepositoryExists(ctx context.Context, repositoryName string) (bool, *errors.AppError) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&models.WorkflowHistory{}).
		Where("repository_name = ?", repositoryName).
		Count(&count).Error; err != nil {
		return false, errors.NewDatabaseQueryError(err).
			WithDetails("failed to check repository existence")
	}

	return count > 0, nil
}