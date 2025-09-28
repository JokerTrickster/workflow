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

// PRCreator handles GitHub Pull Request creation for completed tasks
type PRCreator struct {
	workingDir string
	repoName   string
}

// NewPRCreator creates a new PR creator instance
func NewPRCreator(workingDir, repoName string) *PRCreator {
	return &PRCreator{
		workingDir: workingDir,
		repoName:   repoName,
	}
}

// CreatePRForCompletedTask creates a GitHub PR for a completed task
func (p *PRCreator) CreatePRForCompletedTask(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse) error {
	log.Printf("Starting PR creation process for task: %s", taskMsg.Tasks)

	// Step 1: Validate environment and repository
	if err := p.validateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	// Step 2: Check if we're in the correct working directory
	workDir := p.workingDir
	if taskMsg.WorkingDir != "" {
		workDir = taskMsg.WorkingDir
	}

	if err := p.changeToWorkingDirectory(workDir); err != nil {
		return fmt.Errorf("failed to change to working directory: %w", err)
	}

	// Step 3: Check Git status and ensure we have changes
	hasChanges, err := p.checkForChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		log.Printf("No changes detected, skipping PR creation")
		return nil
	}

	// Step 4: Ensure we're on a feature branch (not main/master)
	branchName, err := p.ensureFeatureBranch(taskMsg)
	if err != nil {
		return fmt.Errorf("failed to ensure feature branch: %w", err)
	}

	// Step 5: Commit changes
	if err := p.commitChanges(taskMsg, result); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Step 6: Push branch to GitHub
	if err := p.pushBranch(branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	// Step 7: Create Pull Request
	prURL, err := p.createGitHubPR(taskMsg, branchName, result)
	if err != nil {
		return fmt.Errorf("failed to create GitHub PR: %w", err)
	}

	log.Printf("Successfully created PR: %s", prURL)
	return nil
}

// validateEnvironment checks if required tools are available
func (p *PRCreator) validateEnvironment() error {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) not found. Install it from https://cli.github.com/")
	}

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found")
	}

	// Check if gh is authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI not authenticated. Run 'gh auth login'")
	}

	return nil
}

// changeToWorkingDirectory changes to the specified working directory
func (p *PRCreator) changeToWorkingDirectory(workDir string) error {
	if workDir == "" {
		return nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", absPath)
	}

	// Change to directory
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	log.Printf("Changed to working directory: %s", absPath)
	return nil
}

// checkForChanges checks if there are uncommitted changes
func (p *PRCreator) checkForChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// ensureFeatureBranch ensures we're on a feature branch, creates one if needed
func (p *PRCreator) ensureFeatureBranch(taskMsg *TaskMessage) (string, error) {
	// Get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(output))

	// If already on a feature branch, use it
	if !isDefaultBranch(currentBranch) {
		log.Printf("Already on feature branch: %s", currentBranch)
		return currentBranch, nil
	}

	// Create a new feature branch
	branchName := p.generateBranchName(taskMsg)

	cmd = exec.Command("git", "checkout", "-b", branchName)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create feature branch %s: %w", branchName, err)
	}

	log.Printf("Created and switched to feature branch: %s", branchName)
	return branchName, nil
}

// isDefaultBranch checks if the branch is a default branch (main/master)
func isDefaultBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

// generateBranchName creates a branch name based on the task
func (p *PRCreator) generateBranchName(taskMsg *TaskMessage) string {
	// Extract task type and description from task content
	taskDescription := strings.ToLower(taskMsg.Tasks)

	// Simple branch name generation
	timestamp := time.Now().Format("0102-1504") // MMDD-HHMM

	// Clean up task description for branch name
	cleaned := strings.ReplaceAll(taskDescription, " ", "-")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")

	// Limit length
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}

	return fmt.Sprintf("task/%s-%s", cleaned, timestamp)
}

// commitChanges commits the changes with a descriptive message
func (p *PRCreator) commitChanges(taskMsg *TaskMessage, result *AITaskResponse) error {
	// Stage all changes
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit message
	commitMsg := p.generateCommitMessage(taskMsg, result)

	// Commit changes
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	log.Printf("Committed changes with message: %s", commitMsg)
	return nil
}

// generateCommitMessage creates a descriptive commit message
func (p *PRCreator) generateCommitMessage(taskMsg *TaskMessage, result *AITaskResponse) string {
	// Base commit message on task content
	taskSummary := taskMsg.Tasks
	if len(taskSummary) > 100 {
		taskSummary = taskSummary[:100] + "..."
	}

	commitMsg := fmt.Sprintf("feat: %s\n\n", taskSummary)

	// Add execution details
	if result != nil {
		commitMsg += fmt.Sprintf("- Execution time: %v\n", result.ExecutionTime)
		commitMsg += fmt.Sprintf("- Provider: %s\n", taskMsg.Provider)
		if len(result.FilesModified) > 0 {
			commitMsg += fmt.Sprintf("- Files modified: %d\n", len(result.FilesModified))
		}
		if result.TokensUsed > 0 {
			commitMsg += fmt.Sprintf("- Tokens used: %d\n", result.TokensUsed)
		}
	}

	commitMsg += "\n🤖 Generated with Claude Code Workflow System\n"
	commitMsg += "Co-Authored-By: Claude <noreply@anthropic.com>"

	return commitMsg
}

// pushBranch pushes the branch to GitHub
func (p *PRCreator) pushBranch(branchName string) error {
	// Push branch with upstream tracking
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s: %w", branchName, err)
	}

	log.Printf("Pushed branch to GitHub: %s", branchName)
	return nil
}

// createGitHubPR creates a Pull Request using GitHub CLI
func (p *PRCreator) createGitHubPR(taskMsg *TaskMessage, branchName string, result *AITaskResponse) (string, error) {
	// Generate PR title and body
	title := p.generatePRTitle(taskMsg)
	body := p.generatePRBody(taskMsg, result)

	// Create PR using gh CLI
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--head", branchName,
		"--base", "main", // or get default branch dynamically
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}

	prURL := strings.TrimSpace(string(output))
	log.Printf("Created GitHub PR: %s", prURL)

	return prURL, nil
}

// generatePRTitle creates a PR title based on the task
func (p *PRCreator) generatePRTitle(taskMsg *TaskMessage) string {
	taskSummary := taskMsg.Tasks
	if len(taskSummary) > 80 {
		taskSummary = taskSummary[:80] + "..."
	}

	return fmt.Sprintf("Task: %s", taskSummary)
}

// generatePRBody creates a detailed PR description
func (p *PRCreator) generatePRBody(taskMsg *TaskMessage, result *AITaskResponse) string {
	body := fmt.Sprintf("## Task Description\n%s\n\n", taskMsg.Tasks)

	body += "## Implementation Details\n"
	body += fmt.Sprintf("- **Repository**: %s\n", taskMsg.RepositoryName)
	body += fmt.Sprintf("- **Provider**: %s\n", taskMsg.Provider)
	body += fmt.Sprintf("- **Interactive**: %t\n", taskMsg.Interactive)

	if taskMsg.WorkingDir != "" {
		body += fmt.Sprintf("- **Working Directory**: %s\n", taskMsg.WorkingDir)
	}

	if result != nil {
		body += "\n## Execution Results\n"
		body += fmt.Sprintf("- **Execution Time**: %v\n", result.ExecutionTime)
		body += fmt.Sprintf("- **Success**: %t\n", result.Success)

		if len(result.FilesModified) > 0 {
			body += fmt.Sprintf("- **Files Modified**: %d\n", len(result.FilesModified))
		}

		if result.TokensUsed > 0 {
			body += fmt.Sprintf("- **Tokens Used**: %d\n", result.TokensUsed)
		}

		if result.Output != "" && len(result.Output) < 1000 {
			body += fmt.Sprintf("\n**Output**:\n```\n%s\n```\n", result.Output)
		}
	}

	body += "\n## Review Checklist\n"
	body += "- [ ] Code changes are correct and complete\n"
	body += "- [ ] Tests pass (if applicable)\n"
	body += "- [ ] Documentation updated (if needed)\n"
	body += "- [ ] No security vulnerabilities introduced\n"
	body += "- [ ] Code follows project standards\n"

	body += "\n---\n"
	body += "🤖 *This PR was automatically created by the Claude Code Workflow System*"

	return body
}