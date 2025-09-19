package claude

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/infrastructure/config"
)

// Client wraps the Anthropic Claude API client
type Client struct {
	client *anthropic.Client
	config *config.ClaudeConfig
}

// NewClient creates a new Claude API client
func NewClient(cfg *config.ClaudeConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))

	return &Client{
		client: &client,
		config: cfg,
	}, nil
}

// CreateConversation sends a message to Claude and returns the response
func (c *Client) CreateConversation(ctx context.Context, req *ConversationRequest) (*ConversationResponse, error) {
	log.Printf("Creating Claude conversation with %d messages", len(req.Messages))

	// For now, implement a simplified version that returns mock response
	// TODO: Implement proper Anthropic SDK integration
	log.Printf("Processing request with system prompt: %s", req.SystemPrompt)
	
	var lastUserMessage string
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			lastUserMessage = msg.Content
		}
		log.Printf("Message [%s]: %s", msg.Role, msg.Content)
	}

	// Simulate processing delay
	time.Sleep(100 * time.Millisecond)

	// Return mock response for now
	mockResponse := fmt.Sprintf("Mock Claude response for: %s", lastUserMessage)
	
	tokenUsage := &TokenUsage{
		InputTokens:  len(lastUserMessage) / 4,  // rough estimate
		OutputTokens: len(mockResponse) / 4,
		TotalTokens:  (len(lastUserMessage) + len(mockResponse)) / 4,
	}

	log.Printf("Mock Claude API call - Input: %d tokens, Output: %d tokens", 
		tokenUsage.InputTokens, tokenUsage.OutputTokens)

	return &ConversationResponse{
		Content:    mockResponse,
		Role:       "assistant",
		TokenUsage: tokenUsage,
		Model:      "claude-3-sonnet-20240229",
		ID:         "mock-response-" + time.Now().Format("20060102150405"),
	}, nil
}

// CreateCodeReview creates a code review using Claude
func (c *Client) CreateCodeReview(ctx context.Context, code string, context string) (*CodeReviewResponse, error) {
	systemPrompt := `You are an expert code reviewer. Analyze the provided code and give constructive feedback focusing on:
1. Code quality and best practices
2. Potential bugs or issues
3. Performance considerations
4. Security concerns
5. Maintainability and readability

Provide your response in JSON format with the following structure:
{
  "overall_score": 1-10,
  "summary": "Brief overall assessment",
  "issues": [
    {
      "type": "bug|performance|security|style",
      "severity": "low|medium|high",
      "line": line_number_or_null,
      "description": "Description of the issue",
      "suggestion": "Suggested improvement"
    }
  ],
  "positive_aspects": ["List of good things about the code"],
  "recommendations": ["Overall recommendations for improvement"]
}`

	messages := []ConversationMessage{
		{
			Role:    string(entities.MessageRoleUser),
			Content: fmt.Sprintf("Context: %s\n\nCode to review:\n```\n%s\n```", context, code),
		},
	}

	req := &ConversationRequest{
		Messages:     messages,
		SystemPrompt: systemPrompt,
	}

	response, err := c.CreateConversation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create code review: %w", err)
	}

	return &CodeReviewResponse{
		Content:    response.Content,
		TokenUsage: response.TokenUsage,
	}, nil
}

// AnalyzeIssue analyzes a GitHub issue using Claude
func (c *Client) AnalyzeIssue(ctx context.Context, issue string, context string) (*IssueAnalysisResponse, error) {
	systemPrompt := `You are an expert software engineer analyzing GitHub issues. Provide detailed analysis focusing on:
1. Problem understanding and root cause analysis
2. Technical complexity assessment
3. Suggested implementation approach
4. Potential risks and considerations
5. Time estimation

Provide your response in JSON format with the following structure:
{
  "complexity": "low|medium|high",
  "category": "bug|feature|enhancement|documentation",
  "estimated_hours": number,
  "summary": "Brief analysis summary",
  "root_cause": "Potential root cause if it's a bug",
  "implementation_approach": "Suggested approach to solve the issue",
  "technical_considerations": ["List of technical aspects to consider"],
  "risks": ["Potential risks or challenges"],
  "dependencies": ["Required dependencies or prerequisites"]
}`

	messages := []ConversationMessage{
		{
			Role:    string(entities.MessageRoleUser),
			Content: fmt.Sprintf("Context: %s\n\nIssue to analyze:\n%s", context, issue),
		},
	}

	req := &ConversationRequest{
		Messages:     messages,
		SystemPrompt: systemPrompt,
	}

	response, err := c.CreateConversation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze issue: %w", err)
	}

	return &IssueAnalysisResponse{
		Content:    response.Content,
		TokenUsage: response.TokenUsage,
	}, nil
}

// Health checks the Claude API connection
func (c *Client) Health(ctx context.Context) error {
	// Simple health check with a minimal request
	messages := []ConversationMessage{
		{
			Role:    string(entities.MessageRoleUser),
			Content: "Hello, respond with 'OK' if you're working.",
		},
	}

	req := &ConversationRequest{
		Messages: messages,
	}

	_, err := c.CreateConversation(ctx, req)
	if err != nil {
		return fmt.Errorf("Claude API health check failed: %w", err)
	}

	return nil
}

// GetTokenUsage returns estimated token usage for a text
func (c *Client) GetTokenUsage(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English text
	return len(text) / 4
}