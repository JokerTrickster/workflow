package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeProvider implements AIProvider interface for Claude CLI
type ClaudeProvider struct {
	APIKey      string
	CLIPath     string
	WorkingDir  string
	Timeout     time.Duration
}

// NewClaudeProvider creates a new Claude provider instance
func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		APIKey:     os.Getenv("CLAUDE_API_KEY"),
		CLIPath:    "claude",  // Assumes claude CLI is in PATH
		Timeout:    30 * time.Minute,
	}
}

// GetProviderName returns the name of the provider
func (c *ClaudeProvider) GetProviderName() string {
	return "claude"
}

// IsConfigured checks if the provider is properly configured
func (c *ClaudeProvider) IsConfigured() bool {
	// Check if API key is set
	if c.APIKey == "" {
		return false
	}

	// Check if CLI is available
	_, err := exec.LookPath(c.CLIPath)
	return err == nil
}

// ExecuteTask executes a task using Claude CLI
func (c *ClaudeProvider) ExecuteTask(ctx context.Context, request *AITaskRequest) (*AITaskResponse, error) {
	startTime := time.Now()

	// Set timeout if specified in request
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	// Prepare working directory
	workingDir := request.WorkingDir
	if workingDir == "" {
		workingDir = c.WorkingDir
	}
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Build Claude CLI command
	args := c.buildClaudeCommand(request)

	// Execute command
	result, err := c.executeCommand(ctx, args, workingDir)
	if err != nil {
		return &AITaskResponse{
			Success:       false,
			Error:         err.Error(),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	return &AITaskResponse{
		Content:       result.Output,
		Success:       result.Success,
		Error:         result.Error,
		TokensUsed:    result.TokensUsed,
		ExecutionTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"provider":     "claude",
			"working_dir":  workingDir,
			"interactive":  request.Interactive,
			"command_used": strings.Join(args, " "),
		},
	}, nil
}

// buildClaudeCommand builds the Claude CLI command arguments
func (c *ClaudeProvider) buildClaudeCommand(request *AITaskRequest) []string {
	args := []string{}

	// Basic claude command
	if request.Interactive {
		args = append(args, "--interactive")
	}

	// Add task content
	if request.Tasks != "" {
		args = append(args, "--message", request.Tasks)
	}

	// Add working directory if specified
	if request.WorkingDir != "" {
		args = append(args, "--working-dir", request.WorkingDir)
	}

	// Add repository context if available
	if request.RepositoryName != "" {
		args = append(args, "--context", fmt.Sprintf("repository:%s", request.RepositoryName))
	}

	// Add continue task flag if needed
	if request.ContinueTask {
		args = append(args, "--continue")
	}

	// Add custom command if specified
	if request.Cmd != "" {
		args = append(args, "--execute", request.Cmd)
	}

	// Add provider-specific options
	if options, ok := request.Options["claude"].(map[string]interface{}); ok {
		for key, value := range options {
			switch key {
			case "model":
				if model, ok := value.(string); ok {
					args = append(args, "--model", model)
				}
			case "temperature":
				if temp, ok := value.(float64); ok {
					args = append(args, "--temperature", fmt.Sprintf("%.2f", temp))
				}
			case "max_tokens":
				if tokens, ok := value.(int); ok {
					args = append(args, "--max-tokens", fmt.Sprintf("%d", tokens))
				}
			}
		}
	}

	return args
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Output     string
	Success    bool
	Error      string
	TokensUsed int
}

// executeCommand executes the Claude CLI command
func (c *ClaudeProvider) executeCommand(ctx context.Context, args []string, workingDir string) (*CommandResult, error) {
	// Create command
	cmd := exec.CommandContext(ctx, c.CLIPath, args...)
	cmd.Dir = workingDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CLAUDE_API_KEY=%s", c.APIKey),
	)

	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	result := &CommandResult{
		Output:  string(output),
		Success: err == nil,
	}

	if err != nil {
		result.Error = err.Error()

		// Check for specific error types
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "command execution timeout"
		}
	}

	// Try to parse tokens used from output if available
	result.TokensUsed = c.parseTokensUsed(result.Output)

	return result, nil
}

// parseTokensUsed attempts to parse token usage from Claude CLI output
func (c *ClaudeProvider) parseTokensUsed(output string) int {
	// Claude CLI might output token usage in various formats
	// This is a placeholder implementation
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "tokens used:") || strings.Contains(line, "Tokens:") {
			// Try to extract number
			parts := strings.Fields(line)
			for i, part := range parts {
				if (strings.Contains(strings.ToLower(part), "token") ||
					strings.Contains(strings.ToLower(part), "usage")) && i+1 < len(parts) {
					var tokens int
					if _, err := fmt.Sscanf(parts[i+1], "%d", &tokens); err == nil {
						return tokens
					}
				}
			}
		}
	}
	return 0
}

// CreateClaudeScript creates a temporary script file for complex operations
func (c *ClaudeProvider) CreateClaudeScript(content string, workingDir string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp(workingDir, "claude-script-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary script file: %w", err)
	}
	defer tmpFile.Close()

	// Write content
	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to write script content: %w", err)
	}

	return tmpFile.Name(), nil
}

// CleanupScript removes temporary script files
func (c *ClaudeProvider) CleanupScript(scriptPath string) error {
	if scriptPath != "" && filepath.Ext(scriptPath) == ".md" {
		return os.Remove(scriptPath)
	}
	return nil
}

// GetClaudeVersion returns the version of Claude CLI
func (c *ClaudeProvider) GetClaudeVersion() (string, error) {
	cmd := exec.Command(c.CLIPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Claude version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateClaudeConfig validates the Claude configuration
func (c *ClaudeProvider) ValidateClaudeConfig() error {
	if !c.IsConfigured() {
		return fmt.Errorf("Claude CLI is not properly configured")
	}

	// Test basic functionality
	cmd := exec.Command(c.CLIPath, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Claude CLI is not working properly: %w", err)
	}

	return nil
}

// init function registers the Claude provider
func init() {
	claudeProvider := NewClaudeProvider()
	GlobalAIProviderFactory.RegisterProvider("claude", claudeProvider)
}