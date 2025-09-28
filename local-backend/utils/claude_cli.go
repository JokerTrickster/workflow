package utils

import (
	"context"
	"fmt"
	"log"
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
	apiKey := os.Getenv("CLAUDE_API_KEY")
	fmt.Printf("[DEBUG] Claude API Key length: %d\n", len(apiKey))
	if apiKey == "" {
		fmt.Println("[ERROR] CLAUDE_API_KEY environment variable is empty")
	}
	return &ClaudeProvider{
		APIKey:     apiKey,
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

	// Prepare working directory - clone repository if needed
	workingDir, err := c.prepareRepositoryWorkspace(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare repository workspace: %w", err)
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

	// For non-interactive mode, use --print flag
	if !request.Interactive {
		args = append(args, "--print")
	}

	// Add task content as the prompt (last argument)
	prompt := request.Tasks

	// Enhance prompt with context information
	if request.RepositoryName != "" {
		prompt = fmt.Sprintf("Repository: %s\n\n%s", request.RepositoryName, prompt)
	}

	if request.Cmd != "" {
		prompt = fmt.Sprintf("%s\n\nAdditional command to consider: %s", prompt, request.Cmd)
	}

	if request.ContinueTask {
		prompt = fmt.Sprintf("%s\n\nNote: This is a continuation of a previous task.", prompt)
	}

	// Add the final prompt as the last argument
	args = append(args, prompt)

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

	// Debug logging
	fmt.Printf("[DEBUG] Executing Claude CLI command:\n")
	fmt.Printf("[DEBUG] Command: %s %s\n", c.CLIPath, strings.Join(args, " "))
	fmt.Printf("[DEBUG] Working directory: %s\n", workingDir)
	fmt.Printf("[DEBUG] API Key present: %t (length: %d)\n", len(c.APIKey) > 0, len(c.APIKey))

	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	result := &CommandResult{
		Output:  string(output),
		Success: err == nil,
	}

	if err != nil {
		result.Error = err.Error()
		fmt.Printf("[DEBUG] Command failed with error: %v\n", err)
		fmt.Printf("[DEBUG] Command output: %s\n", string(output))

		// Check for specific error types
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "command execution timeout"
		}
	} else {
		fmt.Printf("[DEBUG] Command succeeded\n")
		fmt.Printf("[DEBUG] Command output: %s\n", string(output))
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

// prepareRepositoryWorkspace clones repository and creates working branch
func (c *ClaudeProvider) prepareRepositoryWorkspace(ctx context.Context, request *AITaskRequest) (string, error) {
	// If no repository name, use fallback directory logic
	if request.RepositoryName == "" {
		workingDir := request.WorkingDir
		if workingDir == "" {
			workingDir = c.WorkingDir
		}
		if workingDir == "" {
			var err error
			workingDir, err = os.Getwd()
			if err != nil {
				return "", fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
		return workingDir, c.ensureWorkingDirectory(workingDir)
	}

	// Create a unique workspace for this task
	timestamp := time.Now().Unix()
	workspaceDir := fmt.Sprintf("/tmp/claude-workspace-%s-%d", request.RepositoryName, timestamp)

	log.Printf("Creating workspace for repository %s at %s", request.RepositoryName, workspaceDir)

	// Clone the repository
	if err := c.cloneRepository(ctx, request.RepositoryName, workspaceDir); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Create and checkout a working branch
	branchName := fmt.Sprintf("claude-task-%d", timestamp)
	if err := c.createWorkingBranch(ctx, workspaceDir, branchName); err != nil {
		log.Printf("Warning: failed to create branch %s: %v", branchName, err)
		// Continue with main branch if branch creation fails
	}

	return workspaceDir, nil
}

// cloneRepository clones a GitHub repository
func (c *ClaudeProvider) cloneRepository(ctx context.Context, repositoryName, targetDir string) error {
	githubURL := fmt.Sprintf("https://github.com/%s.git", repositoryName)

	log.Printf("Cloning repository %s to %s", githubURL, targetDir)

	// Remove existing directory if it exists
	if err := os.RemoveAll(targetDir); err != nil {
		log.Printf("Failed to remove existing directory: %v", err)
	}

	// Clone the repository
	cmd := exec.CommandContext(ctx, "git", "clone", githubURL, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone %s: %w\nOutput: %s", githubURL, err, string(output))
	}

	log.Printf("Successfully cloned repository %s", repositoryName)
	return nil
}

// createWorkingBranch creates and checks out a new branch
func (c *ClaudeProvider) createWorkingBranch(ctx context.Context, workingDir, branchName string) error {
	log.Printf("Creating working branch: %s", branchName)

	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	log.Printf("Successfully created and switched to branch: %s", branchName)
	return nil
}

// ensureWorkingDirectory creates the working directory if it doesn't exist
func (c *ClaudeProvider) ensureWorkingDirectory(workingDir string) error {
	// Check if directory exists
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		log.Printf("Working directory %s does not exist, creating it", workingDir)
		// Create directory with appropriate permissions
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			return fmt.Errorf("failed to create working directory %s: %w", workingDir, err)
		}
		log.Printf("Created working directory: %s", workingDir)
	} else if err != nil {
		return fmt.Errorf("failed to check working directory %s: %w", workingDir, err)
	} else {
		// Directory exists, check if it's actually a directory
		info, err := os.Stat(workingDir)
		if err != nil {
			return fmt.Errorf("failed to stat working directory %s: %w", workingDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", workingDir)
		}
		log.Printf("Working directory exists: %s", workingDir)
	}

	return nil
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

// RegisterClaudeProvider registers the Claude provider (call after loading env vars)
func RegisterClaudeProvider() {
	fmt.Println("[DEBUG] Registering Claude provider")
	claudeProvider := NewClaudeProvider()
	GlobalAIProviderFactory.RegisterProvider("claude", claudeProvider)
	fmt.Printf("[DEBUG] Claude provider registered, configured: %t\n", claudeProvider.IsConfigured())
}