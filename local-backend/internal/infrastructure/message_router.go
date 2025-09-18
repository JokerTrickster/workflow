package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// MessageRouter handles routing of different message types to appropriate processors
type MessageRouter struct {
	requestService  domain.RequestService
	requestRepo     domain.RequestRepository
}

// NewMessageRouter creates a new message router
func NewMessageRouter(requestService domain.RequestService, requestRepo domain.RequestRepository) *MessageRouter {
	return &MessageRouter{
		requestService: requestService,
		requestRepo:    requestRepo,
	}
}

// RouteMessage routes incoming messages to appropriate handlers based on message type
func (m *MessageRouter) RouteMessage(ctx context.Context, message *domain.Message) error {
	log.Printf("Routing message: ID=%s, Type=%s", message.ID, message.Type)

	switch message.Type {
	case domain.MessageTypeWorkRequest:
		return m.handleWorkRequest(ctx, message)
	case domain.MessageTypeCancellation:
		return m.handleCancellation(ctx, message)
	default:
		return fmt.Errorf("unsupported message type: %s", message.Type)
	}
}

// handleWorkRequest processes work request messages
func (m *MessageRouter) handleWorkRequest(ctx context.Context, message *domain.Message) error {
	log.Printf("Processing work request: %s", message.ID)

	// Validate work request payload
	if err := m.validateWorkRequestPayload(message.Payload); err != nil {
		return domain.NewInvalidRequestError(fmt.Sprintf("invalid work request payload: %v", err))
	}

	// Create request entity
	request, err := m.requestService.CreateRequest(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Process the request asynchronously or synchronously based on configuration
	if err := m.requestService.ProcessRequest(ctx, request); err != nil {
		log.Printf("Failed to process request %s: %v", request.ID, err)
		return fmt.Errorf("failed to process request: %w", err)
	}

	log.Printf("Successfully processed work request: %s", message.ID)
	return nil
}

// handleCancellation processes cancellation messages
func (m *MessageRouter) handleCancellation(ctx context.Context, message *domain.Message) error {
	log.Printf("Processing cancellation request: %s", message.ID)

	// Extract request ID to cancel
	requestIDToCancel, ok := message.Payload["request_id"].(string)
	if !ok || requestIDToCancel == "" {
		return domain.NewInvalidRequestError("cancellation message must include 'request_id' field")
	}

	// Cancel the request
	if err := m.requestService.CancelRequest(ctx, requestIDToCancel); err != nil {
		if err == domain.ErrRequestNotFound {
			log.Printf("Request %s not found for cancellation", requestIDToCancel)
			return nil // Consider this a successful operation
		}
		if err == domain.ErrRequestCannotBeCancelled {
			log.Printf("Request %s cannot be cancelled (already completed)", requestIDToCancel)
			return nil // Consider this a successful operation
		}
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	log.Printf("Successfully cancelled request: %s", requestIDToCancel)
	return nil
}

// validateWorkRequestPayload validates the structure of work request payload
func (m *MessageRouter) validateWorkRequestPayload(payload map[string]interface{}) error {
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

	// Validate content length
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

// GetMessageHandlerFunc returns a function compatible with domain.MessageHandler
func (m *MessageRouter) GetMessageHandlerFunc() domain.MessageHandler {
	return func(ctx context.Context, message *domain.Message) error {
		return m.RouteMessage(ctx, message)
	}
}

// MessagePublisher handles publishing messages to RabbitMQ (for testing/admin purposes)
type MessagePublisher struct {
	config  *RabbitMQConfig
	channel *amqp.Channel
}

// NewMessagePublisher creates a new message publisher
func NewMessagePublisher(config *RabbitMQConfig) (*MessagePublisher, error) {
	connection, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue (idempotent)
	_, err = channel.QueueDeclare(
		config.QueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &MessagePublisher{
		config:  config,
		channel: channel,
	}, nil
}

// PublishWorkRequest publishes a work request message
func (m *MessagePublisher) PublishWorkRequest(messageID, code, task string, contextID *string) error {
	payload := map[string]interface{}{
		"code": code,
		"task": task,
	}

	message := domain.NewMessage(messageID, domain.MessageTypeWorkRequest, payload)
	if contextID != nil {
		message.SetContextID(*contextID)
	}

	return m.publishMessage(message)
}

// PublishCancellation publishes a cancellation message
func (m *MessagePublisher) PublishCancellation(messageID, requestIDToCancel string) error {
	payload := map[string]interface{}{
		"request_id": requestIDToCancel,
	}

	message := domain.NewMessage(messageID, domain.MessageTypeCancellation, payload)
	return m.publishMessage(message)
}

// publishMessage publishes a message to the queue
func (m *MessagePublisher) publishMessage(message *domain.Message) error {
	// Convert to JSON
	messageData := map[string]interface{}{
		"id":      message.ID,
		"type":    string(message.Type),
		"payload": message.Payload,
	}
	
	if contextID := message.GetContextID(); contextID != "" {
		messageData["context_id"] = contextID
	}

	body, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to queue
	err = m.channel.Publish(
		m.config.Exchange,   // exchange
		m.config.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message: %s", message.ID)
	return nil
}

// Close closes the publisher connection
func (m *MessagePublisher) Close() error {
	if m.channel != nil {
		return m.channel.Close()
	}
	return nil
}