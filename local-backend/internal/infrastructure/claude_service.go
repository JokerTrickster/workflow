package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// SimplifiedClaudeService implements Claude API using HTTP client directly
type SimplifiedClaudeService struct {
	config      *ClaudeConfig
	contextRepo domain.ProcessingContextRepository
	httpClient  *http.Client
}

// NewSimplifiedClaudeService creates a new simplified Claude service
func NewSimplifiedClaudeService(config *ClaudeConfig, contextRepo domain.ProcessingContextRepository) (*SimplifiedClaudeService, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return &SimplifiedClaudeService{
		config:      config,
		contextRepo: contextRepo,
		httpClient:  httpClient,
	}, nil
}

// ClaudeMessage represents a message in the conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the request to Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// SendRequest sends a request to Claude API with context management
func (c *SimplifiedClaudeService) SendRequest(ctx context.Context, content string, contextID string) (string, error) {
	log.Printf("Sending request to Claude API (contextID: %s)", contextID)

	// Get or create processing context
	processingContext, err := c.getOrCreateContext(ctx, contextID)
	if err != nil {
		return "", fmt.Errorf("failed to get processing context: %w", err)
	}

	// Send request with context
	response, err := c.SendRequestWithContext(ctx, content, processingContext)
	if err != nil {
		return "", err
	}

	// Update context with the conversation
	c.updateContext(ctx, processingContext, content, response)

	return response, nil
}

// SendRequestWithContext sends a request using existing context
func (c *SimplifiedClaudeService) SendRequestWithContext(ctx context.Context, content string, processingContext *domain.ProcessingContext) (string, error) {
	// Prepare messages
	var messages []ClaudeMessage

	// Add conversation history
	for _, contextMsg := range processingContext.GetMessages() {
		messages = append(messages, ClaudeMessage{
			Role:    contextMsg.Role,
			Content: contextMsg.Content,
		})
	}

	// Add current message
	messages = append(messages, ClaudeMessage{
		Role:    "user",
		Content: content,
	})

	// Create request
	request := ClaudeRequest{
		Model:     c.config.Model,
		MaxTokens: c.config.MaxTokens,
		Messages:  messages,
	}

	// Send HTTP request
	response, err := c.sendHTTPRequest(ctx, request)
	if err != nil {
		return "", domain.NewServiceUnavailableError("Claude API", err)
	}

	// Extract text response
	if len(response.Content) == 0 || response.Content[0].Type != "text" {
		return "", fmt.Errorf("unexpected response format from Claude API")
	}

	responseText := response.Content[0].Text
	log.Printf("Received response from Claude API (%d chars, %d input tokens, %d output tokens)", 
		len(responseText), response.Usage.InputTokens, response.Usage.OutputTokens)

	return responseText, nil
}

// sendHTTPRequest sends HTTP request to Claude API
func (c *SimplifiedClaudeService) sendHTTPRequest(ctx context.Context, request ClaudeRequest) (*ClaudeResponse, error) {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var claudeResponse ClaudeResponse
	if err := json.Unmarshal(responseBody, &claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &claudeResponse, nil
}

// getOrCreateContext retrieves existing context or creates a new one
func (c *SimplifiedClaudeService) getOrCreateContext(ctx context.Context, contextID string) (*domain.ProcessingContext, error) {
	if contextID == "" {
		// Generate a new context ID
		contextID = fmt.Sprintf("ctx_%d", time.Now().UnixNano())
	}

	// Try to get existing context
	processingContext, err := c.contextRepo.GetByID(ctx, contextID)
	if err == domain.ErrContextNotFound {
		// Create new context
		processingContext = domain.NewProcessingContext(contextID)
		if err := c.contextRepo.Create(ctx, processingContext); err != nil {
			return nil, fmt.Errorf("failed to create processing context: %w", err)
		}
		log.Printf("Created new processing context: %s", contextID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve processing context: %w", err)
	} else {
		// Update last used time
		processingContext.LastUsedAt = time.Now()
	}

	return processingContext, nil
}

// updateContext updates the processing context with new conversation
func (c *SimplifiedClaudeService) updateContext(ctx context.Context, processingContext *domain.ProcessingContext, userMessage, assistantMessage string) {
	// Add messages to context
	processingContext.AddUserMessage(userMessage)
	processingContext.AddAssistantMessage(assistantMessage)

	// Update in repository
	if err := c.contextRepo.Update(ctx, processingContext); err != nil {
		log.Printf("Warning: failed to update processing context %s: %v", processingContext.ID, err)
	}
}

// Health performs a health check on the Claude API service
func (c *SimplifiedClaudeService) Health(ctx context.Context) error {
	// Simple health check by sending a minimal request
	request := ClaudeRequest{
		Model:     c.config.Model,
		MaxTokens: 10,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: "ping",
			},
		},
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.sendHTTPRequest(timeoutCtx, request)
	if err != nil {
		return fmt.Errorf("Claude API health check failed: %w", err)
	}

	return nil
}

// GetStats returns Claude service statistics
func (c *SimplifiedClaudeService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"model":       c.config.Model,
		"max_tokens":  c.config.MaxTokens,
		"timeout":     c.config.Timeout,
		"api_key_set": c.config.APIKey != "",
		"api_endpoint": "https://api.anthropic.com/v1/messages",
	}
}

// CleanupExpiredContexts removes old contexts that haven't been used recently
func (c *SimplifiedClaudeService) CleanupExpiredContexts(ctx context.Context, maxAge time.Duration) error {
	log.Printf("Cleaning up expired contexts (max age: %v)", maxAge)
	
	err := c.contextRepo.CleanupExpiredContexts(ctx, int64(maxAge.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to cleanup expired contexts: %w", err)
	}
	
	log.Println("Context cleanup completed")
	return nil
}

// ValidateConfiguration validates the Claude service configuration
func (c *SimplifiedClaudeService) ValidateConfiguration() error {
	if c.config.APIKey == "" {
		return fmt.Errorf("Claude API key is required")
	}
	
	if c.config.Model == "" {
		return fmt.Errorf("Claude model is required")
	}
	
	if c.config.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}
	
	if c.config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}