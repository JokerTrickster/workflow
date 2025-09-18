package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// MessageProcessorImpl implements the MessageProcessor interface
type MessageProcessorImpl struct {
	requestService domain.RequestService
	contextService domain.ContextService
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(
	requestService domain.RequestService,
	contextService domain.ContextService,
) domain.MessageProcessor {
	return &MessageProcessorImpl{
		requestService: requestService,
		contextService: contextService,
	}
}

// ProcessMessage handles incoming message processing
func (m *MessageProcessorImpl) ProcessMessage(ctx context.Context, message *domain.Message) error {
	log.Printf("Processing message: ID=%s, Type=%s", message.ID, message.Type)

	switch message.Type {
	case domain.MessageTypeWorkRequest:
		return m.processWorkRequest(ctx, message)
	case domain.MessageTypeCancellation:
		return m.processCancellation(ctx, message)
	default:
		return fmt.Errorf("unsupported message type: %s", message.Type)
	}
}

// processWorkRequest processes work request messages
func (m *MessageProcessorImpl) processWorkRequest(ctx context.Context, message *domain.Message) error {
	log.Printf("Processing work request: %s", message.ID)

	// Validate message payload
	if err := m.validateWorkRequestPayload(message.Payload); err != nil {
		return domain.NewInvalidRequestError(fmt.Sprintf("invalid payload: %v", err))
	}

	// Create request through request service
	request, err := m.requestService.CreateRequest(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Process the request
	if err := m.requestService.ProcessRequest(ctx, request); err != nil {
		log.Printf("Failed to process request %s: %v", request.ID, err)
		return fmt.Errorf("failed to process request: %w", err)
	}

	log.Printf("Successfully processed work request: %s", message.ID)
	return nil
}

// processCancellation processes cancellation messages
func (m *MessageProcessorImpl) processCancellation(ctx context.Context, message *domain.Message) error {
	log.Printf("Processing cancellation: %s", message.ID)

	// Extract request ID to cancel
	requestID, ok := message.Payload["request_id"].(string)
	if !ok || requestID == "" {
		return domain.NewInvalidRequestError("cancellation message must include 'request_id' field")
	}

	// Cancel the request through request service
	if err := m.requestService.CancelRequest(ctx, requestID); err != nil {
		if err == domain.ErrRequestNotFound {
			log.Printf("Request %s not found for cancellation", requestID)
			return nil // Consider this a successful operation
		}
		if err == domain.ErrRequestCannotBeCancelled {
			log.Printf("Request %s cannot be cancelled (already completed)", requestID)
			return nil // Consider this a successful operation
		}
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	log.Printf("Successfully processed cancellation for request: %s", requestID)
	return nil
}

// validateWorkRequestPayload validates the structure of work request payload
func (m *MessageProcessorImpl) validateWorkRequestPayload(payload map[string]interface{}) error {
	// Check required fields
	if _, ok := payload["code"]; !ok {
		return fmt.Errorf("missing required field: code")
	}
	
	if _, ok := payload["task"]; !ok {
		return fmt.Errorf("missing required field: task")
	}

	// Validate field types
	if _, ok := payload["code"].(string); !ok {
		return fmt.Errorf("field 'code' must be a string")
	}
	
	if _, ok := payload["task"].(string); !ok {
		return fmt.Errorf("field 'task' must be a string")
	}

	// Validate content
	code := payload["code"].(string)
	task := payload["task"].(string)
	
	if len(code) == 0 {
		return fmt.Errorf("field 'code' cannot be empty")
	}
	
	if len(task) == 0 {
		return fmt.Errorf("field 'task' cannot be empty")
	}
	
	if len(code) > 100000 { // 100KB limit
		return fmt.Errorf("field 'code' exceeds maximum length of 100KB")
	}

	return nil
}