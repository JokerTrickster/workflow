package repositories

import (
	"context"
	"local-backend-server/internal/domain/entities"
	"time"
)

// RequestRepository defines the interface for request data access
type RequestRepository interface {
	// Create creates a new request
	Create(ctx context.Context, request *entities.Request) error
	
	// GetByID retrieves a request by its ID
	GetByID(ctx context.Context, id string) (*entities.Request, error)
	
	// GetBySessionID retrieves all requests for a session
	GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Request, error)
	
	// GetByStatus retrieves requests by status
	GetByStatus(ctx context.Context, status entities.RequestStatus) ([]*entities.Request, error)
	
	// GetPendingRequests retrieves all pending requests
	GetPendingRequests(ctx context.Context) ([]*entities.Request, error)
	
	// GetProcessingRequests retrieves all processing requests
	GetProcessingRequests(ctx context.Context) ([]*entities.Request, error)
	
	// Update updates an existing request
	Update(ctx context.Context, request *entities.Request) error
	
	// Delete deletes a request by ID
	Delete(ctx context.Context, id string) error
	
	// GetByTypeAndStatus retrieves requests by type and status
	GetByTypeAndStatus(ctx context.Context, requestType entities.RequestType, status entities.RequestStatus) ([]*entities.Request, error)
	
	// GetRequestsCreatedAfter retrieves requests created after a specific time
	GetRequestsCreatedAfter(ctx context.Context, after time.Time) ([]*entities.Request, error)
	
	// GetRequestsWithTimeout retrieves requests that may have timed out
	GetRequestsWithTimeout(ctx context.Context, timeoutDuration time.Duration) ([]*entities.Request, error)
	
	// CountByStatus returns the count of requests by status
	CountByStatus(ctx context.Context, status entities.RequestStatus) (int, error)
	
	// GetRequestMetrics retrieves request processing metrics
	GetRequestMetrics(ctx context.Context, from, to time.Time) (*RequestMetrics, error)
}

// RequestMetrics holds request processing statistics
type RequestMetrics struct {
	TotalRequests     int     `json:"total_requests"`
	CompletedRequests int     `json:"completed_requests"`
	FailedRequests    int     `json:"failed_requests"`
	AverageTimeMs     float64 `json:"average_time_ms"`
	SuccessRate       float64 `json:"success_rate"`
}