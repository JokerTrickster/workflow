package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/infrastructure/queue"
	"local-backend-server/internal/interfaces"
)

// AtomicQueueService handles atomic queue publishing and database persistence
type AtomicQueueService struct {
	db        *gorm.DB
	publisher interfaces.MessagePublisher
}

// NewAtomicQueueService creates a new atomic queue service
func NewAtomicQueueService(db *gorm.DB, publisher interfaces.MessagePublisher) *AtomicQueueService {
	return &AtomicQueueService{
		db:        db,
		publisher: publisher,
	}
}

// PublishWithHistory atomically publishes a message to the queue and saves it to the database
func (s *AtomicQueueService) PublishWithHistory(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	// Generate UUID for request
	requestID := uuid.New().String()
	
	var response *PublishResponse
	
	// Start database transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create workflow history record
		history := &models.WorkflowHistory{
			RequestID:      requestID,
			Status:         models.WorkflowStatusPending,
			Tasks:          req.Tasks,
			RepositoryName: req.RepositoryName,
			Interactive:    req.Interactive,
			ContinueTask:   req.ContinueTask,
			CreatedAt:      time.Now(),
		}
		
		// Set optional fields if provided
		if req.WorkingDir != "" {
			history.WorkingDir = &req.WorkingDir
		}
		if req.ClaudeCmd != "" {
			history.ClaudeCmd = &req.ClaudeCmd
		}
		
		// Save to database
		if err := tx.Create(history).Error; err != nil {
			return fmt.Errorf("failed to create workflow history: %w", err)
		}
		
		// Create queue message
		workflowMessage := &queue.WorkflowMessage{
			Type:      req.MessageType,
			ID:        requestID,
			SessionID: req.SessionID,
			Payload:   req.Payload,
			Timestamp: time.Now(),
		}
		
		// Attempt to publish to queue
		// Note: RabbitMQ operates outside of database transactions
		// If this fails, the database transaction will be rolled back
		if s.publisher != nil {
			if err := s.publisher.PublishMessage(workflowMessage); err != nil {
				return fmt.Errorf("failed to publish message to queue: %w", err)
			}
		}
		
		// Prepare response
		response = &PublishResponse{
			RequestID: requestID,
			Status:    models.WorkflowStatusPending,
			Message:   "Task queued successfully",
			CreatedAt: history.CreatedAt,
		}
		
		return nil
	})
	
	if err != nil {
		// If transaction failed, we need to handle potential queue message cleanup
		// In a more sophisticated implementation, we might use a compensating transaction
		// or implement an outbox pattern for guaranteed consistency
		return nil, fmt.Errorf("atomic operation failed: %w", err)
	}
	
	return response, nil
}

// GetWorkflowHistory retrieves workflow history by request ID
func (s *AtomicQueueService) GetWorkflowHistory(requestID string) (*models.WorkflowHistory, error) {
	var history models.WorkflowHistory
	err := s.db.Where("request_id = ?", requestID).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (s *AtomicQueueService) UpdateWorkflowStatus(requestID string, status string, result *string, error *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == models.WorkflowStatusCompleted || status == models.WorkflowStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}
	
	if result != nil {
		updates["result"] = result
	}
	
	if error != nil {
		updates["error"] = error
	}
	
	return s.db.Model(&models.WorkflowHistory{}).
		Where("request_id = ?", requestID).
		Updates(updates).Error
}

// PublishRequest represents a request to publish a message atomically
type PublishRequest struct {
	Tasks          string
	RepositoryName string
	WorkingDir     string
	Interactive    bool
	ClaudeCmd      string
	ContinueTask   bool
	MessageType    string
	SessionID      string
	Payload        map[string]interface{}
}

// PublishResponse represents the response from an atomic publish operation
type PublishResponse struct {
	RequestID string    `json:"request_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}