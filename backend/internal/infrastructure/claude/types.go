package claude

// ConversationRequest represents a request to Claude API
type ConversationRequest struct {
	Messages     []ConversationMessage `json:"messages"`
	SystemPrompt string                `json:"system_prompt,omitempty"`
	MaxTokens    int                   `json:"max_tokens,omitempty"`
	Temperature  float64               `json:"temperature,omitempty"`
}

// ConversationMessage represents a single message in a conversation
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ConversationResponse represents a response from Claude API
type ConversationResponse struct {
	Content    string      `json:"content"`
	Role       string      `json:"role"`
	TokenUsage *TokenUsage `json:"token_usage"`
	Model      string      `json:"model"`
	ID         string      `json:"id"`
}

// TokenUsage represents token consumption information
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// CodeReviewResponse represents the response from a code review request
type CodeReviewResponse struct {
	Content    string      `json:"content"`
	TokenUsage *TokenUsage `json:"token_usage"`
}

// IssueAnalysisResponse represents the response from an issue analysis request
type IssueAnalysisResponse struct {
	Content    string      `json:"content"`
	TokenUsage *TokenUsage `json:"token_usage"`
}

// WorkflowResponse represents a structured response for workflow processing
type WorkflowResponse struct {
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	Data       map[string]interface{} `json:"data"`
	Actions    []string               `json:"actions,omitempty"`
	TokenUsage *TokenUsage            `json:"token_usage"`
}

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	System      string            `json:"system"`
	Template    string            `json:"template"`
	Variables   map[string]string `json:"variables"`
}

// ConversationContext represents the context for maintaining conversations
type ConversationContext struct {
	ID               string                  `json:"id"`
	SessionID        string                  `json:"session_id"`
	RequestID        string                  `json:"request_id"`
	Messages         []ConversationMessage   `json:"messages"`
	SystemPrompt     string                  `json:"system_prompt"`
	Metadata         map[string]interface{}  `json:"metadata"`
	TotalTokenUsage  *TokenUsage             `json:"total_token_usage"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
}