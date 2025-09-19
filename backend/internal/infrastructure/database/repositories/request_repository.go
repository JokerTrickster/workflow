package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"local-backend-server/internal/infrastructure/database/models"
)

// requestRepository implements the RequestRepository interface
type requestRepository struct {
	db *gorm.DB
}

// NewRequestRepository creates a new request repository
func NewRequestRepository(db *gorm.DB) repositories.RequestRepository {
	return &requestRepository{db: db}
}

// Create creates a new request
func (r *requestRepository) Create(ctx context.Context, request *entities.Request) error {
	model := &models.Request{
		ID:               request.ID,
		SessionID:        request.SessionID,
		Type:             string(request.Type),
		Status:           string(request.Status),
		Input:            models.JSON(request.Input),
		Output:           models.JSON(request.Output),
		Error:            request.Error,
		ProcessingTimeMs: request.ProcessingTimeMs,
		CreatedAt:        request.CreatedAt,
		UpdatedAt:        request.UpdatedAt,
		StartedAt:        request.StartedAt,
		CompletedAt:      request.CompletedAt,
	}
	
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a request by its ID
func (r *requestRepository) GetByID(ctx context.Context, id string) (*entities.Request, error) {
	var model models.Request
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// GetBySessionID retrieves all requests for a session
func (r *requestRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Request, error) {
	var models []models.Request
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	requests := make([]*entities.Request, len(models))
	for i, model := range models {
		requests[i] = r.modelToEntity(&model)
	}
	
	return requests, nil
}

// GetByStatus retrieves requests by status
func (r *requestRepository) GetByStatus(ctx context.Context, status entities.RequestStatus) ([]*entities.Request, error) {
	var models []models.Request
	err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	requests := make([]*entities.Request, len(models))
	for i, model := range models {
		requests[i] = r.modelToEntity(&model)
	}
	
	return requests, nil
}

// GetPendingRequests retrieves all pending requests
func (r *requestRepository) GetPendingRequests(ctx context.Context) ([]*entities.Request, error) {
	return r.GetByStatus(ctx, entities.RequestStatusPending)
}

// GetProcessingRequests retrieves all processing requests
func (r *requestRepository) GetProcessingRequests(ctx context.Context) ([]*entities.Request, error) {
	return r.GetByStatus(ctx, entities.RequestStatusProcessing)
}

// Update updates an existing request
func (r *requestRepository) Update(ctx context.Context, request *entities.Request) error {
	model := &models.Request{
		ID:               request.ID,
		SessionID:        request.SessionID,
		Type:             string(request.Type),
		Status:           string(request.Status),
		Input:            models.JSON(request.Input),
		Output:           models.JSON(request.Output),
		Error:            request.Error,
		ProcessingTimeMs: request.ProcessingTimeMs,
		CreatedAt:        request.CreatedAt,
		UpdatedAt:        request.UpdatedAt,
		StartedAt:        request.StartedAt,
		CompletedAt:      request.CompletedAt,
	}
	
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete deletes a request by ID
func (r *requestRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Request{}, "id = ?", id).Error
}

// GetByTypeAndStatus retrieves requests by type and status
func (r *requestRepository) GetByTypeAndStatus(ctx context.Context, requestType entities.RequestType, status entities.RequestStatus) ([]*entities.Request, error) {
	var models []models.Request
	err := r.db.WithContext(ctx).
		Where("type = ? AND status = ?", string(requestType), string(status)).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	requests := make([]*entities.Request, len(models))
	for i, model := range models {
		requests[i] = r.modelToEntity(&model)
	}
	
	return requests, nil
}

// GetRequestsCreatedAfter retrieves requests created after a specific time
func (r *requestRepository) GetRequestsCreatedAfter(ctx context.Context, after time.Time) ([]*entities.Request, error) {
	var models []models.Request
	err := r.db.WithContext(ctx).
		Where("created_at > ?", after).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	requests := make([]*entities.Request, len(models))
	for i, model := range models {
		requests[i] = r.modelToEntity(&model)
	}
	
	return requests, nil
}

// GetRequestsWithTimeout retrieves requests that may have timed out
func (r *requestRepository) GetRequestsWithTimeout(ctx context.Context, timeoutDuration time.Duration) ([]*entities.Request, error) {
	timeoutThreshold := time.Now().Add(-timeoutDuration)
	
	var models []models.Request
	err := r.db.WithContext(ctx).
		Where("status = ? AND started_at IS NOT NULL AND started_at < ?", 
			string(entities.RequestStatusProcessing), timeoutThreshold).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	requests := make([]*entities.Request, len(models))
	for i, model := range models {
		requests[i] = r.modelToEntity(&model)
	}
	
	return requests, nil
}

// CountByStatus returns the count of requests by status
func (r *requestRepository) CountByStatus(ctx context.Context, status entities.RequestStatus) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Request{}).
		Where("status = ?", string(status)).
		Count(&count).Error
	
	return int(count), err
}

// GetRequestMetrics retrieves request processing metrics
func (r *requestRepository) GetRequestMetrics(ctx context.Context, from, to time.Time) (*repositories.RequestMetrics, error) {
	var metrics repositories.RequestMetrics
	
	// Total requests
	var totalCount int64
	err := r.db.WithContext(ctx).
		Model(&models.Request{}).
		Where("created_at BETWEEN ? AND ?", from, to).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	metrics.TotalRequests = int(totalCount)
	
	// Completed requests
	var completedCount int64
	err = r.db.WithContext(ctx).
		Model(&models.Request{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", 
			string(entities.RequestStatusCompleted), from, to).
		Count(&completedCount).Error
	if err != nil {
		return nil, err
	}
	metrics.CompletedRequests = int(completedCount)
	
	// Failed requests
	var failedCount int64
	err = r.db.WithContext(ctx).
		Model(&models.Request{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", 
			string(entities.RequestStatusFailed), from, to).
		Count(&failedCount).Error
	if err != nil {
		return nil, err
	}
	metrics.FailedRequests = int(failedCount)
	
	// Average processing time
	var avgTime float64
	err = r.db.WithContext(ctx).
		Model(&models.Request{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", 
			string(entities.RequestStatusCompleted), from, to).
		Select("AVG(processing_time_ms)").
		Scan(&avgTime).Error
	if err != nil {
		return nil, err
	}
	metrics.AverageTimeMs = avgTime
	
	// Success rate
	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = float64(metrics.CompletedRequests) / float64(metrics.TotalRequests) * 100
	}
	
	return &metrics, nil
}

// modelToEntity converts a database model to domain entity
func (r *requestRepository) modelToEntity(model *models.Request) *entities.Request {
	return &entities.Request{
		ID:               model.ID,
		SessionID:        model.SessionID,
		Type:             entities.RequestType(model.Type),
		Status:           entities.RequestStatus(model.Status),
		Input:            map[string]interface{}(model.Input),
		Output:           map[string]interface{}(model.Output),
		Error:            model.Error,
		ProcessingTimeMs: model.ProcessingTimeMs,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		StartedAt:        model.StartedAt,
		CompletedAt:      model.CompletedAt,
	}
}