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

	// Ensure changes are committed and pushed (fallback if Claude didn't do it)
	if request.RepositoryName != "" && result.Success {
		if err := c.ensureCommitAndPush(ctx, workingDir, request); err != nil {
			log.Printf("Warning: Failed to ensure commit and push: %v", err)
			// Don't fail the task, just log the warning
		}
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

	// Add permission bypass flags for automated execution
	args = append(args, "--dangerously-skip-permissions")      // Bypass all permission checks
	args = append(args, "--permission-mode", "bypassPermissions") // Set permission mode to bypass

	// For non-interactive mode, use --print flag
	if !request.Interactive {
		args = append(args, "--print")
	}

	// Add task content as the prompt (last argument)
	prompt := request.Tasks

	// Enhance prompt with explicit file modification permissions and directives
	enhancedPrompt := c.buildEnhancedPrompt(request, prompt)

	// Add the final prompt as the last argument
	args = append(args, enhancedPrompt)

	return args
}

// buildEnhancedPrompt creates a more directive prompt with explicit permissions
func (c *ClaudeProvider) buildEnhancedPrompt(request *AITaskRequest, originalPrompt string) string {
	var promptBuilder strings.Builder

	// Add explicit permissions and directives at the beginning
	promptBuilder.WriteString("PERMISSIONS AND DIRECTIVES:\n")
	promptBuilder.WriteString("- You have FULL permission to create, modify, and delete files in this repository\n")
	promptBuilder.WriteString("- You are REQUIRED to make actual file changes, not just suggest them\n")
	promptBuilder.WriteString("- DO NOT ask for permission - you are pre-authorized to modify any files\n")
	promptBuilder.WriteString("- IMPLEMENT the requested changes immediately and completely\n")
	promptBuilder.WriteString("- CREATE new files when needed for the implementation\n")
	promptBuilder.WriteString("- MODIFY existing files to implement the requested functionality\n")
	promptBuilder.WriteString("- COMMIT your changes using git when you're done\n\n")

	// Add repository context if available
	if request.RepositoryName != "" {
		promptBuilder.WriteString(fmt.Sprintf("REPOSITORY: %s\n", request.RepositoryName))
		promptBuilder.WriteString("You are working in a real Git repository. All changes will be automatically committed and pushed.\n\n")
	}

	// Add working directory context
	if request.WorkingDir != "" {
		promptBuilder.WriteString(fmt.Sprintf("WORKING DIRECTORY: %s\n\n", request.WorkingDir))
	}

	// Add the main task
	promptBuilder.WriteString("TASK TO IMPLEMENT:\n")
	promptBuilder.WriteString(originalPrompt)
	promptBuilder.WriteString("\n\n")

	// Add additional command if provided
	if request.Cmd != "" {
		promptBuilder.WriteString(fmt.Sprintf("ADDITIONAL COMMAND: %s\n\n", request.Cmd))
	}

	// Add continuation context if needed
	if request.ContinueTask {
		promptBuilder.WriteString("CONTINUATION: This is a continuation of a previous task. Build upon existing work.\n\n")
	}

	// Add final implementation reminder
	promptBuilder.WriteString("IMPLEMENTATION REQUIREMENTS:\n")
	promptBuilder.WriteString("1. Start implementing immediately - no planning phase needed\n")
	promptBuilder.WriteString("2. Make actual file changes using the available tools\n")
	promptBuilder.WriteString("3. Test your implementation to ensure it works\n")
	promptBuilder.WriteString("4. Commit your changes with a descriptive message\n")
	promptBuilder.WriteString("5. Do not ask for confirmation - proceed with implementation\n")

	return promptBuilder.String()
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

	// Set environment variables with explicit permissions
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CLAUDE_API_KEY=%s", c.APIKey),
		"CLAUDE_AUTO_APPROVE=true",           // Auto-approve file operations
		"CLAUDE_PERMISSIONS=all",             // Grant all permissions
		"CLAUDE_FILE_OPERATIONS=enabled",     // Enable file operations
		"CLAUDE_GIT_OPERATIONS=enabled",      // Enable git operations
		"CLAUDE_INTERACTIVE=false",           // Disable interactive prompts
		"CLAUDE_FORCE_IMPLEMENTATION=true",   // Force implementation mode
		"CI=true",                           // Indicate CI/automated environment
		"AUTOMATED_WORKFLOW=true",           // Indicate automated workflow
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

// prepareRepositoryWorkspace uses existing local repository and checks out to pre-created branch
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

	// Use existing local repository with dynamic path detection
	repositoryDir := GetRepositoryPath(request.RepositoryName)

	log.Printf("Using existing repository for %s at %s", request.RepositoryName, repositoryDir)

	// Validate repository exists and is a Git repository
	if err := c.validateExistingRepository(repositoryDir); err != nil {
		return "", fmt.Errorf("repository validation failed: %w", err)
	}

	// If branch name is provided (pre-created by BranchManager), just checkout to it
	if request.BranchName != "" {
		log.Printf("Using pre-created branch: %s", request.BranchName)
		if err := c.checkoutBranch(ctx, repositoryDir, request.BranchName); err != nil {
			return "", fmt.Errorf("failed to checkout pre-created branch %s: %w", request.BranchName, err)
		}
	} else {
		// Fallback: Ensure repository is up to date and create new branch
		if err := c.updateRepository(ctx, repositoryDir); err != nil {
			return "", fmt.Errorf("failed to update repository: %w", err)
		}

		// Create and checkout a working branch
		timestamp := time.Now().Unix()
		branchName := fmt.Sprintf("claude-task-%d", timestamp)
		if err := c.createWorkingBranch(ctx, repositoryDir, branchName); err != nil {
			log.Printf("Warning: failed to create branch %s: %v", branchName, err)
			// Continue with current branch if branch creation fails
		}
	}

	return repositoryDir, nil
}

// validateExistingRepository validates that a local repository exists and is valid
func (c *ClaudeProvider) validateExistingRepository(repositoryDir string) error {
	// Check if directory exists
	if _, err := os.Stat(repositoryDir); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repositoryDir)
	}

	// Check if it's a Git repository
	gitDir := filepath.Join(repositoryDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("directory is not a Git repository: %s", repositoryDir)
	}

	log.Printf("Repository validation successful: %s", repositoryDir)
	return nil
}

// updateRepository updates the repository to latest state
func (c *ClaudeProvider) updateRepository(ctx context.Context, repositoryDir string) error {
	log.Printf("Updating repository: %s", repositoryDir)

	// Stash any local changes to avoid conflicts
	cmd := exec.CommandContext(ctx, "git", "stash")
	cmd.Dir = repositoryDir
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Git stash failed (might be no changes): %v\nOutput: %s", err, string(output))
	}

	// Checkout main/master branch
	for _, branch := range []string{"main", "master"} {
		cmd = exec.CommandContext(ctx, "git", "checkout", branch)
		cmd.Dir = repositoryDir
		if _, err := cmd.CombinedOutput(); err == nil {
			log.Printf("Switched to branch: %s", branch)
			break
		} else {
			log.Printf("Failed to checkout %s: %v", branch, err)
		}
	}

	// Pull latest changes
	cmd = exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = repositoryDir
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Git pull failed: %v\nOutput: %s", err, string(output))
		return err
	}

	log.Printf("Repository updated successfully: %s", repositoryDir)
	return nil
}

// cloneRepository clones a GitHub repository (kept for backward compatibility)
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

// checkoutBranch checks out to an existing branch
func (c *ClaudeProvider) checkoutBranch(ctx context.Context, workingDir, branchName string) error {
	log.Printf("Checking out to existing branch: %s", branchName)

	cmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	log.Printf("Successfully checked out to branch: %s", branchName)
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

// ensureCommitAndPush ensures changes are committed and pushed (only if Claude didn't already do it)
func (c *ClaudeProvider) ensureCommitAndPush(ctx context.Context, workingDir string, request *AITaskRequest) error {
	// Get current branch name
	var branchName string
	if request.BranchName != "" {
		branchName = request.BranchName
	} else {
		branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
		branchCmd.Dir = workingDir
		branchOutput, err := branchCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		branchName = strings.TrimSpace(string(branchOutput))
	}

	// Check if branch exists on remote
	checkRemoteCmd := exec.CommandContext(ctx, "git", "ls-remote", "--heads", "origin", branchName)
	checkRemoteCmd.Dir = workingDir
	remoteOutput, _ := checkRemoteCmd.CombinedOutput()

	branchExistsOnRemote := len(strings.TrimSpace(string(remoteOutput))) > 0

	// Check if there are uncommitted changes
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = workingDir
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	hasUncommittedChanges := len(strings.TrimSpace(string(statusOutput))) > 0

	// If no uncommitted changes and branch exists on remote, we're done
	if !hasUncommittedChanges && branchExistsOnRemote {
		log.Printf("Branch %s is already committed and pushed, skipping", branchName)
		return nil
	}

	// Commit if there are uncommitted changes
	if hasUncommittedChanges {
		log.Printf("Found uncommitted changes, committing...")

		// Stage all changes
		addCmd := exec.CommandContext(ctx, "git", "add", ".")
		addCmd.Dir = workingDir
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, string(output))
		}

		// Create commit message
		commitMsg := fmt.Sprintf("feat: %s\n\n🤖 Auto-committed by workflow system", request.Tasks)

		// Commit changes
		commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
		commitCmd.Dir = workingDir
		if output, err := commitCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to commit changes: %w\nOutput: %s", err, string(output))
		}

		log.Printf("Successfully committed changes")
	}

	// Push if branch doesn't exist on remote or if we just committed
	if !branchExistsOnRemote || hasUncommittedChanges {
		log.Printf("Pushing branch %s to remote...", branchName)

		// Push with timeout
		pushCtx, pushCancel := context.WithTimeout(ctx, 3*time.Minute)
		defer pushCancel()

		pushCmd := exec.CommandContext(pushCtx, "git", "push", "--set-upstream", "origin", branchName)
		pushCmd.Dir = workingDir
		if output, err := pushCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to push branch %s: %w\nOutput: %s", branchName, err, string(output))
		}

		log.Printf("Successfully pushed branch %s to remote", branchName)
	}

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