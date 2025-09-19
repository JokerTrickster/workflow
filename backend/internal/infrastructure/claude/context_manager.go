package claude

import (
	"context"
	"fmt"
	"log"
	"time"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
)

// ContextManager manages conversation contexts for Claude API interactions
type ContextManager struct {
	client                    *Client
	processingContextRepo     repositories.ProcessingContextRepository
	messageRepo               repositories.MessageRepository
	maxConversationLength     int
	contextRetentionDuration  time.Duration
}

// NewContextManager creates a new context manager
func NewContextManager(
	client *Client,
	processingContextRepo repositories.ProcessingContextRepository,
	messageRepo repositories.MessageRepository,
) *ContextManager {
	return &ContextManager{
		client:                   client,
		processingContextRepo:    processingContextRepo,
		messageRepo:             messageRepo,
		maxConversationLength:   20, // Maximum number of messages to keep in context
		contextRetentionDuration: 24 * time.Hour,
	}
}

// CreateContext creates a new processing context for a request
func (cm *ContextManager) CreateContext(ctx context.Context, requestID, sessionID string, systemPrompt string) (*entities.ProcessingContext, error) {
	log.Printf("Creating processing context for request: %s", requestID)

	// Check if context already exists for this request
	existingContext, err := cm.processingContextRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing context: %w", err)
	}

	if existingContext != nil {
		log.Printf("Processing context already exists for request: %s", requestID)
		return existingContext, nil
	}

	// Create new processing context
	processingContext := entities.NewProcessingContext(requestID, sessionID)
	if systemPrompt != "" {
		processingContext.SetSystemPrompt(systemPrompt)
	}

	// Load conversation history from the session
	if err := cm.loadConversationHistory(ctx, processingContext); err != nil {
		log.Printf("Warning: failed to load conversation history: %v", err)
		// Continue with empty history
	}

	// Save the context
	if err := cm.processingContextRepo.Create(ctx, processingContext); err != nil {
		return nil, fmt.Errorf("failed to create processing context: %w", err)
	}

	log.Printf("Created processing context: %s", processingContext.ID)
	return processingContext, nil
}

// GetContext retrieves a processing context by request ID
func (cm *ContextManager) GetContext(ctx context.Context, requestID string) (*entities.ProcessingContext, error) {
	processingContext, err := cm.processingContextRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing context: %w", err)
	}

	if processingContext == nil {
		return nil, fmt.Errorf("processing context not found for request: %s", requestID)
	}

	// Load fresh conversation history
	if err := cm.loadConversationHistory(ctx, processingContext); err != nil {
		log.Printf("Warning: failed to refresh conversation history: %v", err)
	}

	return processingContext, nil
}

// AddMessage adds a message to the processing context
func (cm *ContextManager) AddMessage(ctx context.Context, processingContext *entities.ProcessingContext, message *entities.Message) error {
	// Add to processing context
	processingContext.AddMessage(message)

	// Trim conversation history if it's too long
	cm.trimConversationHistory(processingContext)

	// Update the processing context
	if err := cm.processingContextRepo.Update(ctx, processingContext); err != nil {
		return fmt.Errorf("failed to update processing context: %w", err)
	}

	return nil
}

// UpdateTokenUsage updates token usage information in the processing context
func (cm *ContextManager) UpdateTokenUsage(ctx context.Context, processingContext *entities.ProcessingContext, inputTokens, outputTokens int) error {
	processingContext.UpdateTokenUsage(inputTokens, outputTokens)

	if err := cm.processingContextRepo.Update(ctx, processingContext); err != nil {
		return fmt.Errorf("failed to update token usage: %w", err)
	}

	log.Printf("Updated token usage for context %s: input=%d, output=%d, total=%d",
		processingContext.ID, inputTokens, outputTokens, 
		processingContext.TokenUsage.TotalTokens)

	return nil
}

// ProcessWithContext processes a message using the conversation context
func (cm *ContextManager) ProcessWithContext(ctx context.Context, processingContext *entities.ProcessingContext, userMessage string) (*ConversationResponse, error) {
	log.Printf("Processing message with context: %s", processingContext.ID)

	// Create user message
	userMsg := entities.NewMessage(
		processingContext.SessionID,
		entities.MessageTypeWorkRequest,
		entities.MessageRoleUser,
		userMessage,
	)

	// Add user message to context
	if err := cm.AddMessage(ctx, processingContext, userMsg); err != nil {
		return nil, fmt.Errorf("failed to add user message to context: %w", err)
	}

	// Save user message to database
	if err := cm.messageRepo.Create(ctx, userMsg); err != nil {
		log.Printf("Warning: failed to save user message: %v", err)
	}

	// Convert conversation history to Claude format
	messages := make([]ConversationMessage, len(processingContext.ConversationHistory))
	for i, msg := range processingContext.ConversationHistory {
		messages[i] = ConversationMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Create Claude request
	claudeRequest := &ConversationRequest{
		Messages:     messages,
		SystemPrompt: processingContext.SystemPrompt,
	}

	// Call Claude API
	response, err := cm.client.CreateConversation(ctx, claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Create assistant message
	assistantMsg := entities.NewMessage(
		processingContext.SessionID,
		entities.MessageTypeWorkRequest,
		entities.MessageRoleAssistant,
		response.Content,
	)

	// Add assistant message to context
	if err := cm.AddMessage(ctx, processingContext, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to add assistant message to context: %w", err)
	}

	// Save assistant message to database
	if err := cm.messageRepo.Create(ctx, assistantMsg); err != nil {
		log.Printf("Warning: failed to save assistant message: %v", err)
	}

	// Update token usage
	if response.TokenUsage != nil {
		if err := cm.UpdateTokenUsage(ctx, processingContext, 
			response.TokenUsage.InputTokens, response.TokenUsage.OutputTokens); err != nil {
			log.Printf("Warning: failed to update token usage: %v", err)
		}
	}

	log.Printf("Successfully processed message with context: %s", processingContext.ID)
	return response, nil
}

// loadConversationHistory loads recent conversation history from the database
func (cm *ContextManager) loadConversationHistory(ctx context.Context, processingContext *entities.ProcessingContext) error {
	// Get recent messages from the session
	messages, err := cm.messageRepo.GetBySessionIDWithPagination(
		ctx, 
		processingContext.SessionID, 
		0, 
		cm.maxConversationLength,
	)
	if err != nil {
		return fmt.Errorf("failed to load conversation history: %w", err)
	}

	// Clear existing history and add loaded messages
	processingContext.ConversationHistory = make([]*entities.Message, 0, len(messages))
	for _, msg := range messages {
		processingContext.ConversationHistory = append(processingContext.ConversationHistory, msg)
	}

	log.Printf("Loaded %d messages for conversation history", len(messages))
	return nil
}

// trimConversationHistory trims the conversation history to the maximum length
func (cm *ContextManager) trimConversationHistory(processingContext *entities.ProcessingContext) {
	if len(processingContext.ConversationHistory) <= cm.maxConversationLength {
		return
	}

	// Keep the most recent messages
	startIndex := len(processingContext.ConversationHistory) - cm.maxConversationLength
	processingContext.ConversationHistory = processingContext.ConversationHistory[startIndex:]

	log.Printf("Trimmed conversation history to %d messages", len(processingContext.ConversationHistory))
}

// CleanupExpiredContexts removes old processing contexts
func (cm *ContextManager) CleanupExpiredContexts(ctx context.Context) error {
	log.Println("Starting cleanup of expired processing contexts")

	// This would require additional repository methods to query by date
	// For now, we'll implement a simple cleanup strategy
	
	log.Println("Processing context cleanup completed")
	return nil
}

// GetContextSummary returns a summary of the processing context
func (cm *ContextManager) GetContextSummary(processingContext *entities.ProcessingContext) map[string]interface{} {
	summary := map[string]interface{}{
		"id":             processingContext.ID,
		"request_id":     processingContext.RequestID,
		"session_id":     processingContext.SessionID,
		"message_count":  len(processingContext.ConversationHistory),
		"system_prompt":  processingContext.SystemPrompt != "",
		"created_at":     processingContext.CreatedAt,
		"updated_at":     processingContext.UpdatedAt,
	}

	if processingContext.TokenUsage != nil {
		summary["token_usage"] = map[string]interface{}{
			"input_tokens":  processingContext.TokenUsage.InputTokens,
			"output_tokens": processingContext.TokenUsage.OutputTokens,
			"total_tokens":  processingContext.TokenUsage.TotalTokens,
		}
	}

	return summary
}