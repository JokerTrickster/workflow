package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"local-backend-server/internal/errors"
	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/infrastructure/queue"
	"local-backend-server/internal/interfaces"
)

// AtomicQueueService handles atomic queue publishing and database persistence
type AtomicQueueService struct {
	db        *gorm.DB
	publisher interfaces.MessagePublisher
	logger    *errors.Logger
}

// NewAtomicQueueService creates a new atomic queue service
func NewAtomicQueueService(db *gorm.DB, publisher interfaces.MessagePublisher) *AtomicQueueService {
	return &AtomicQueueService{
		db:        db,
		publisher: publisher,
		logger:    errors.NewLogger(),
	}
}

// PublishWithHistory atomically publishes a message to the queue and saves it to the database
func (s *AtomicQueueService) PublishWithHistory(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	// Validate input
	if err := s.validatePublishRequest(req); err != nil {
		s.logger.ErrorWithContext(err, "Invalid publish request", map[string]interface{}{
			"repository": req.RepositoryName,
			"tasks":      req.Tasks,
		})
		return nil, err
	}

	// Generate UUID for request
	requestID := uuid.New().String()
	logger := s.logger.WithRequestID(requestID)
	
	logger.InfoWithContext("Starting atomic publish operation", map[string]interface{}{
		"repository":    req.RepositoryName,
		"message_type":  req.MessageType,
		"session_id":    req.SessionID,
		"interactive":   req.Interactive,
	})
	
	var response *PublishResponse
	
	// Check database connection health
	if err := s.checkDatabaseHealth(); err != nil {
		logger.Error(err, "Database health check failed")
		return nil, err
	}

	// Check queue connection health if publisher exists
	if s.publisher != nil {
		if err := s.checkQueueHealth(); err != nil {
			logger.Error(err, "Queue health check failed")
			return nil, err
		}
	}
	
	// Start database transaction with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create workflow history record
		history := &models.WorkflowHistory{
			RequestID:      requestID,
			Status:         models.WorkflowStatusPending,
			Tasks:          req.Tasks,
			RepositoryName: req.RepositoryName,
			Interactive:    req.Interactive,
			ContinueTask:   req.ContinueTask,
			CreatedAt:      time.Now().UTC(),
		}
		
		// Set optional fields if provided
		if req.WorkingDir != "" {
			history.WorkingDir = &req.WorkingDir
		}
		if req.ClaudeCmd != "" {
			history.ClaudeCmd = &req.ClaudeCmd
		}
		
		// Save to database with detailed error handling
		if err := tx.Create(history).Error; err != nil {
			dbErr := s.analyzeDatabaseError(err)
			logger.ErrorWithContext(dbErr, "Failed to create workflow history", map[string]interface{}{
				"request_id": requestID,
				"repository": req.RepositoryName,
			})
			return dbErr
		}
		
		logger.InfoWithContext("Workflow history created", map[string]interface{}{
			"request_id": requestID,
			"status":     history.Status,
		})
		
		// Create queue message
		workflowMessage := &queue.WorkflowMessage{
			Type:      req.MessageType,
			ID:        requestID,
			SessionID: req.SessionID,
			Payload:   req.Payload,
			Timestamp: time.Now().UTC(),
		}
		
		// Attempt to publish to queue if publisher is available
		if s.publisher != nil {
			if err := s.publishWithRetry(workflowMessage, requestID); err != nil {
				logger.ErrorWithContext(err, "Failed to publish message to queue", map[string]interface{}{
					"request_id":   requestID,
					"message_type": req.MessageType,
					"session_id":   req.SessionID,
				})
				return err
			}
			
			logger.InfoWithContext("Message published to queue", map[string]interface{}{
				"request_id":   requestID,
				"message_type": req.MessageType,
			})
		} else {
			logger.WarnWithContext("No queue publisher available, continuing without queue", map[string]interface{}{
				"request_id": requestID,
			})
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
		logger.ErrorWithContext(err, "Atomic operation failed", map[string]interface{}{
			"request_id": requestID,
			"repository": req.RepositoryName,
		})
		
		// Handle specific error types
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		
		// Handle context timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewDatabaseTimeoutError(err).
				WithRequestID(requestID).
				WithContext("operation", "atomic_publish").
				WithContext("repository", req.RepositoryName)
		}
		
		// Generic internal error
		return nil, errors.NewInternalError(err).
			WithRequestID(requestID).
			WithContext("operation", "atomic_publish").
			WithContext("repository", req.RepositoryName)
	}
	
	logger.InfoWithContext("Atomic publish operation completed successfully", map[string]interface{}{
		"request_id": requestID,
		"status":     response.Status,
	})
	
	return response, nil
}

// GetWorkflowHistory retrieves workflow history by request ID
func (s *AtomicQueueService) GetWorkflowHistory(requestID string) (*models.WorkflowHistory, error) {
	if requestID == "" {
		return nil, errors.NewValidationError("Request ID is required").
			WithContext("field", "request_id")
	}

	var history models.WorkflowHistory
	err := s.db.Where("request_id = ?", requestID).First(&history).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("workflow", requestID)
		}
		
		dbErr := s.analyzeDatabaseError(err)
		s.logger.ErrorWithContext(dbErr, "Failed to retrieve workflow history", map[string]interface{}{
			"request_id": requestID,
		})
		return nil, dbErr
	}
	
	return &history, nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (s *AtomicQueueService) UpdateWorkflowStatus(requestID string, status string, result *string, errorMsg *string) error {
	if requestID == "" {
		return errors.NewValidationError("Request ID is required").
			WithContext("field", "request_id")
	}
	
	if status == "" {
		return errors.NewValidationError("Status is required").
			WithContext("field", "status")
	}

	logger := s.logger.WithRequestID(requestID)
	
	logger.InfoWithContext("Updating workflow status", map[string]interface{}{
		"request_id": requestID,
		"status":     status,
		"has_result": result != nil,
		"has_error":  errorMsg != nil,
	})

	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == models.WorkflowStatusCompleted || status == models.WorkflowStatusFailed {
		now := time.Now().UTC()
		updates["completed_at"] = &now
	}
	
	if result != nil {
		updates["result"] = result
	}
	
	if errorMsg != nil {
		updates["error"] = errorMsg
	}
	
	err := s.db.Model(&models.WorkflowHistory{}).
		Where("request_id = ?", requestID).
		Updates(updates).Error
		
	if err != nil {
		dbErr := s.analyzeDatabaseError(err)
		logger.ErrorWithContext(dbErr, "Failed to update workflow status", map[string]interface{}{
			"request_id": requestID,
			"status":     status,
		})
		return dbErr
	}
	
	logger.InfoWithContext("Workflow status updated successfully", map[string]interface{}{
		"request_id": requestID,
		"new_status": status,
	})
	
	return nil
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

// validatePublishRequest validates the publish request
func (s *AtomicQueueService) validatePublishRequest(req PublishRequest) error {
	if req.Tasks == "" {
		return errors.NewMissingFieldError("tasks")
	}
	
	if req.RepositoryName == "" {
		return errors.NewMissingFieldError("repository_name")
	}
	
	if req.MessageType == "" {
		return errors.NewMissingFieldError("message_type")
	}
	
	if req.SessionID == "" {
		return errors.NewMissingFieldError("session_id")
	}
	
	// Validate repository name format (basic validation)
	if len(req.RepositoryName) > 255 {
		return errors.NewInvalidInputError("repository_name", req.RepositoryName).
			WithDetails("Repository name must be 255 characters or less")
	}
	
	// Validate tasks length
	if len(req.Tasks) > 10000 {
		return errors.NewInvalidInputError("tasks", len(req.Tasks)).
			WithDetails("Tasks description must be 10,000 characters or less")
	}
	
	return nil
}

// checkDatabaseHealth checks if the database connection is healthy
func (s *AtomicQueueService) checkDatabaseHealth() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return errors.NewDatabaseConnectionError(err).
			WithDetails("Failed to get database instance")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return errors.NewDatabaseConnectionError(err).
			WithDetails("Database ping failed")
	}
	
	return nil
}

// checkQueueHealth checks if the queue connection is healthy
func (s *AtomicQueueService) checkQueueHealth() error {
	// For now, we assume if publisher exists, it's healthy
	// In a more sophisticated implementation, we might add a health check method to the publisher interface
	if s.publisher == nil {
		return errors.NewQueueConnectionError(nil).
			WithDetails("Queue publisher is not available")
	}
	return nil
}

// publishWithRetry attempts to publish a message with retry logic
func (s *AtomicQueueService) publishWithRetry(message *queue.WorkflowMessage, requestID string) error {
	const maxRetries = 3
	const retryDelay = time.Second * 2
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.publisher.PublishMessage(message)
		if err == nil {
			return nil
		}
		
		s.logger.WarnWithContext("Queue publish attempt failed", map[string]interface{}{
			"request_id": requestID,
			"attempt":    attempt,
			"max_retries": maxRetries,
			"error":      err.Error(),
		})
		
		// Don't sleep on the last attempt
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}
	
	return errors.NewQueuePublishError(fmt.Errorf("failed to publish after %d attempts", maxRetries)).
		WithRequestID(requestID).
		WithContext("attempts", maxRetries)
}

// analyzeDatabaseError analyzes database errors and returns appropriate AppError
func (s *AtomicQueueService) analyzeDatabaseError(err error) *errors.AppError {
	if err == nil {
		return nil
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Check for specific database errors
	if err == gorm.ErrRecordNotFound {
		return errors.NewNotFoundError("record", "").
			WithDetails("The requested record was not found")
	}
	
	// Check for connection errors
	if strings.Contains(errStr, "connection") || 
	   strings.Contains(errStr, "connect") ||
	   strings.Contains(errStr, "network") {
		return errors.NewDatabaseConnectionError(err)
	}
	
	// Check for timeout errors
	if strings.Contains(errStr, "timeout") || 
	   strings.Contains(errStr, "deadline exceeded") {
		return errors.NewDatabaseTimeoutError(err)
	}
	
	// Check for constraint violations
	if strings.Contains(errStr, "constraint") ||
	   strings.Contains(errStr, "unique") ||
	   strings.Contains(errStr, "foreign key") ||
	   strings.Contains(errStr, "check constraint") {
		return errors.NewDatabaseConstraintError(err)
	}
	
	// Check for driver errors
	if err == sql.ErrNoRows {
		return errors.NewNotFoundError("record", "").
			WithDetails("No matching records found")
	}
	
	if err == sql.ErrConnDone {
		return errors.NewDatabaseConnectionError(err).
			WithDetails("Database connection is closed")
	}
	
	// Generic database error
	return errors.NewDatabaseQueryError(err)
}