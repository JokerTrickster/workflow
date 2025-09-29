package utils

import (
	"context"
	"time"
)

// AIProvider represents the common interface for AI providers
type AIProvider interface {
	// ExecuteTask executes a task using the AI provider
	ExecuteTask(ctx context.Context, request *AITaskRequest) (*AITaskResponse, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string

	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
}

// AITaskRequest represents a request to an AI provider
type AITaskRequest struct {
	// Task description or prompt
	Tasks string `json:"tasks"`

	// Repository information
	RepositoryName string `json:"repository_name"`
	WorkingDir     string `json:"working_dir"`

	// Execution options
	Interactive  bool   `json:"interactive"`
	Cmd          string `json:"cmd"`
	ContinueTask bool   `json:"continue_task"`

	// Request metadata
	RequestID string `json:"request_id"`
	Timeout   time.Duration `json:"timeout"`

	// Provider-specific options
	Options map[string]interface{} `json:"options,omitempty"`
}

// AITaskResponse represents a response from an AI provider
type AITaskResponse struct {
	// Response content
	Content string `json:"content"`
	Output  string `json:"output,omitempty"`

	// Execution results
	Success       bool     `json:"success"`
	Error         string   `json:"error,omitempty"`
	FilesModified []string `json:"files_modified,omitempty"`

	// Usage information
	TokensUsed      int           `json:"tokens_used,omitempty"`
	ExecutionTime   time.Duration `json:"execution_time"`

	// Provider-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AIProviderFactory creates AI provider instances
type AIProviderFactory struct {
	providers map[string]AIProvider
}

// NewAIProviderFactory creates a new provider factory
func NewAIProviderFactory() *AIProviderFactory {
	return &AIProviderFactory{
		providers: make(map[string]AIProvider),
	}
}

// RegisterProvider registers an AI provider
func (f *AIProviderFactory) RegisterProvider(name string, provider AIProvider) {
	f.providers[name] = provider
}

// GetProvider returns an AI provider by name
func (f *AIProviderFactory) GetProvider(name string) (AIProvider, bool) {
	provider, exists := f.providers[name]
	return provider, exists
}

// GetAvailableProviders returns a list of available provider names
func (f *AIProviderFactory) GetAvailableProviders() []string {
	var names []string
	for name, provider := range f.providers {
		if provider.IsConfigured() {
			names = append(names, name)
		}
	}
	return names
}

// Global factory instance
var GlobalAIProviderFactory = NewAIProviderFactory()