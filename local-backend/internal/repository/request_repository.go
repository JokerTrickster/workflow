package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// RequestRepositoryImpl implements the RequestRepository interface
type RequestRepositoryImpl struct {
	db *gorm.DB
}

// NewRequestRepository creates a new request repository
func NewRequestRepository(db *gorm.DB) domain.RequestRepository {
	return &RequestRepositoryImpl{
		db: db,
	}
}

// Create saves a new request
func (r *RequestRepositoryImpl) Create(ctx context.Context, request *domain.Request) error {
	if err := r.db.WithContext(ctx).Create(request).Error; err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	return nil
}

// GetByID retrieves a request by its ID
func (r *RequestRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Request, error) {
	var request domain.Request
	err := r.db.WithContext(ctx).First(&request, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRequestNotFound
		}
		return nil, fmt.Errorf("failed to get request by ID: %w", err)
	}
	return &request, nil
}

// GetByMessageID retrieves a request by message ID
func (r *RequestRepositoryImpl) GetByMessageID(ctx context.Context, messageID string) (*domain.Request, error) {
	var request domain.Request
	err := r.db.WithContext(ctx).First(&request, "message_id = ?", messageID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRequestNotFound
		}
		return nil, fmt.Errorf("failed to get request by message ID: %w", err)
	}
	return &request, nil
}

// GetByContextID retrieves all requests for a context
func (r *RequestRepositoryImpl) GetByContextID(ctx context.Context, contextID string) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("context_id = ?", contextID).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get requests by context ID: %w", err)
	}
	return requests, nil
}

// GetByStatus retrieves requests by status
func (r *RequestRepositoryImpl) GetByStatus(ctx context.Context, status domain.RequestStatus) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get requests by status: %w", err)
	}
	return requests, nil
}

// Update updates an existing request
func (r *RequestRepositoryImpl) Update(ctx context.Context, request *domain.Request) error {
	result := r.db.WithContext(ctx).Save(request)
	if result.Error != nil {
		return fmt.Errorf("failed to update request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrRequestNotFound
	}
	return nil
}

// Delete removes a request (soft delete recommended)
func (r *RequestRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Request{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrRequestNotFound
	}
	return nil
}

// GetPendingRequests retrieves all pending requests
func (r *RequestRepositoryImpl) GetPendingRequests(ctx context.Context) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.RequestStatusPending).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	return requests, nil
}

// GetProcessingRequests retrieves all currently processing requests
func (r *RequestRepositoryImpl) GetProcessingRequests(ctx context.Context) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.RequestStatusProcessing).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get processing requests: %w", err)
	}
	return requests, nil
}

// Additional utility methods

// GetRequestsWithStatus retrieves requests with multiple statuses
func (r *RequestRepositoryImpl) GetRequestsWithStatus(ctx context.Context, statuses []domain.RequestStatus) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("status IN ?", statuses).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get requests with statuses: %w", err)
	}
	return requests, nil
}

// CountByStatus counts requests by status
func (r *RequestRepositoryImpl) CountByStatus(ctx context.Context, status domain.RequestStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Request{}).
		Where("status = ?", status).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count requests by status: %w", err)
	}
	return count, nil
}

// GetRequestsCreatedBetween retrieves requests created within a time range
func (r *RequestRepositoryImpl) GetRequestsCreatedBetween(ctx context.Context, start, end time.Time) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get requests by date range: %w", err)
	}
	return requests, nil
}