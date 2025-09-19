package claude

import (
	"context"
	"fmt"
	"log"
	"math"
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
		client: client,
		config: cfg,
	}, nil
}

// CreateConversation sends a message to Claude and returns the response
func (c *Client) CreateConversation(ctx context.Context, req *ConversationRequest) (*ConversationResponse, error) {
	log.Printf("Creating Claude conversation with %d messages", len(req.Messages))

	// Convert messages to Anthropic format
	messages := make([]anthropic.MessageParam, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = anthropic.MessageParam{
			Role:    anthropic.MessageRole(msg.Role),
			Content: anthropic.TextBlockParam{Type: anthropic.TextBlockTypeText, Text: msg.Content},
		}
	}

	// Create the request
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3Sonnet20240229,
		MaxTokens: anthropic.Int(int64(c.config.MaxTokens)),
		Messages:  anthropic.F(messages),
	}

	// Add system message if provided
	if req.SystemPrompt != "" {
		params.System = anthropic.F(req.SystemPrompt)
	}

	// Make the API call with retry logic
	var response *anthropic.Message
	var err error

	for attempt := 1; attempt <= 3; attempt++ {
		response, err = c.client.Messages.New(ctx, params)
		if err == nil {
			break
		}

		if attempt < 3 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Printf("Claude API call failed (attempt %d/3), retrying in %v: %v", attempt, backoff, err)
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API after 3 attempts: %w", err)
	}

	// Parse the response
	if len(response.Content) == 0 {
		return nil, fmt.Errorf("Claude API returned empty response")
	}

	// Extract text content
	var content string
	for _, block := range response.Content {
		if textBlock, ok := block.AsTextBlock(); ok {
			content += textBlock.Text
		}
	}

	if content == "" {
		return nil, fmt.Errorf("Claude API response contains no text content")
	}

	// Calculate token usage
	tokenUsage := &TokenUsage{
		InputTokens:  int(response.Usage.InputTokens),
		OutputTokens: int(response.Usage.OutputTokens),
		TotalTokens:  int(response.Usage.InputTokens + response.Usage.OutputTokens),
	}

	log.Printf("Claude API call successful - Input: %d tokens, Output: %d tokens", 
		tokenUsage.InputTokens, tokenUsage.OutputTokens)

	return &ConversationResponse{
		Content:    content,
		Role:       string(response.Role),
		TokenUsage: tokenUsage,
		Model:      string(response.Model),
		ID:         response.ID,
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