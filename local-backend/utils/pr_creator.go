package utils

import (
	"context"
	"fmt"
	"log"
	"main/features/github/model/response"
	"main/features/github/service"
	"os"
	"os/exec"
	"strings"
	"time"
)

// PRCreator handles GitHub Pull Request creation for completed tasks
type PRCreator struct {
	workingDir    string
	repoName      string
	githubService *service.GitHubService
	githubToken   string
}

// NewPRCreator creates a new PR creator instance
func NewPRCreator(workingDir, repoName string) *PRCreator {
	return &PRCreator{
		workingDir:    workingDir,
		repoName:      repoName,
		githubService: service.NewGitHubService(),
		githubToken:   os.Getenv("GITHUB_TOKEN"),
	}
}

// CreatePRForCompletedTask creates a GitHub PR for a completed task
func (p *PRCreator) CreatePRForCompletedTask(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse) error {
	return p.CreatePRForCompletedTaskWithIssue(ctx, taskMsg, result, "")
}

// CreatePRForCompletedTaskWithIssue creates a GitHub PR for a completed task with issue linking
func (p *PRCreator) CreatePRForCompletedTaskWithIssue(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse, issueNumber string) error {
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
		log.Printf("No changes detected, skipping PR creation for task: %s", taskMsg.Tasks)
		return nil
	}

	// Step 4: Get current branch name
	branchName, err := p.getCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	log.Printf("Current branch: %s", branchName)

	// Step 5: Ensure changes are committed and pushed
	if err := p.ensureChangesPushed(branchName); err != nil {
		return fmt.Errorf("failed to ensure changes are pushed: %w", err)
	}

	// Step 6: Create GitHub Pull Request
	prURL, err := p.createGitHubPRWithIssue(taskMsg, branchName, result, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to create GitHub PR: %w", err)
	}

	log.Printf("Successfully created PR for task '%s': %s", taskMsg.Tasks, prURL)
	return nil
}

// validateEnvironment checks if all necessary tools are available
func (p *PRCreator) validateEnvironment() error {
	// Check if we have git
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found: %w", err)
	}

	// Check if we have gh CLI (optional, we have API fallback)
	if _, err := exec.LookPath("gh"); err != nil {
		log.Printf("Warning: GitHub CLI not found, will use API only")
	}

	return nil
}

// changeToWorkingDirectory changes to the specified working directory
func (p *PRCreator) changeToWorkingDirectory(workDir string) error {
	if workDir == "" {
		return fmt.Errorf("working directory not specified")
	}

	// Check if directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", workDir)
	}

	// Change to directory
	if err := os.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", workDir, err)
	}

	// Update working directory reference
	p.workingDir = workDir

	log.Printf("Changed to working directory: %s", workDir)
	return nil
}

// checkForChanges verifies if there are any Git changes to commit
func (p *PRCreator) checkForChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = p.workingDir

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	hasChanges := len(strings.TrimSpace(string(output))) > 0
	log.Printf("Git status check - has changes: %t", hasChanges)

	return hasChanges, nil
}

// getCurrentBranch gets the current Git branch name
func (p *PRCreator) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = p.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branchName := strings.TrimSpace(string(output))
	if branchName == "" {
		return "", fmt.Errorf("unable to determine current branch")
	}

	return branchName, nil
}

// ensureChangesPushed makes sure all changes are committed and pushed
func (p *PRCreator) ensureChangesPushed(branchName string) error {
	// Check if there are uncommitted changes
	hasUncommitted, err := p.checkForChanges()
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if hasUncommitted {
		log.Printf("Found uncommitted changes, but they should have been committed by the task worker")
		// Don't auto-commit here, let the task worker handle this
	}

	// Check if branch exists on remote
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branchName)
	cmd.Dir = p.workingDir

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check remote branch: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		log.Printf("Branch %s doesn't exist on remote, pushing...", branchName)
		return p.pushBranch(branchName)
	}

	log.Printf("Branch %s already exists on remote", branchName)
	return nil
}

// pushBranch pushes the current branch to GitHub
func (p *PRCreator) pushBranch(branchName string) error {
	log.Printf("Pushing branch to GitHub: %s", branchName)

	cmd := exec.Command("git", "push", "--set-upstream", "origin", branchName)
	cmd.Dir = p.workingDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s: %w", branchName, err)
	}

	log.Printf("Pushed branch to GitHub: %s", branchName)
	return nil
}

// createGitHubPRWithIssue creates a Pull Request using GitHub API or CLI fallback
func (p *PRCreator) createGitHubPRWithIssue(taskMsg *TaskMessage, branchName string, result *AITaskResponse, issueNumber string) (string, error) {
	// Try GitHub API first
	if p.githubToken != "" && p.githubService != nil {
		prURL, err := p.createPRWithAPI(taskMsg, branchName, result, issueNumber)
		if err == nil {
			return prURL, nil
		}
		log.Printf("GitHub API PR creation failed, falling back to CLI: %v", err)
	}

	// Fallback to GitHub CLI
	return p.createPRWithCLI(taskMsg, branchName, result, issueNumber)
}

// createPRWithAPI creates a Pull Request using the GitHub API
func (p *PRCreator) createPRWithAPI(taskMsg *TaskMessage, branchName string, result *AITaskResponse, issueNumber string) (string, error) {
	// Parse repository name
	parts := strings.Split(taskMsg.RepositoryName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository name format: %s", taskMsg.RepositoryName)
	}
	owner, repo := parts[0], parts[1]

	// Get default branch
	defaultBranch, err := p.getDefaultBranch()
	if err != nil {
		log.Printf("Warning: failed to get default branch, using 'main': %v", err)
		defaultBranch = "main"
	}

	// Generate PR title and body
	title := p.generatePRTitle(taskMsg)
	body := p.generatePRBodyWithIssue(taskMsg, result, issueNumber)

	// Create PR request
	prReq := &response.CreatePullRequestRequest{
		Title: title,
		Body:  &body,
		Head:  branchName,
		Base:  defaultBranch,
	}

	// Create the pull request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pr, err := p.githubService.CreatePullRequest(ctx, p.githubToken, owner, repo, prReq)
	if err != nil {
		return "", fmt.Errorf("failed to create PR with API: %w", err)
	}

	log.Printf("Created GitHub PR #%d with API: %s", pr.Number, pr.HTMLURL)
	return pr.HTMLURL, nil
}

// createPRWithCLI creates a Pull Request using GitHub CLI
func (p *PRCreator) createPRWithCLI(taskMsg *TaskMessage, branchName string, result *AITaskResponse, issueNumber string) (string, error) {
	// Generate PR title and body
	title := p.generatePRTitle(taskMsg)
	body := p.generatePRBodyWithIssue(taskMsg, result, issueNumber)

	// Get default branch
	defaultBranch, err := p.getDefaultBranch()
	if err != nil {
		log.Printf("Warning: failed to get default branch, using 'main': %v", err)
		defaultBranch = "main"
	}

	// Create PR using gh CLI
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--head", branchName,
		"--base", defaultBranch,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create PR with CLI: %w", err)
	}

	prURL := strings.TrimSpace(string(output))
	log.Printf("Created GitHub PR with CLI: %s", prURL)

	return prURL, nil
}

// getDefaultBranch determines the default branch of the repository
func (p *PRCreator) getDefaultBranch() (string, error) {
	// Try to get default branch from Git
	cmd := exec.Command("git", "remote", "show", "origin")
	cmd.Dir = p.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "HEAD branch:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not determine default branch")
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
	return p.generatePRBodyWithIssue(taskMsg, result, "")
}

// generatePRBodyWithIssue creates a detailed PR description with issue linking
func (p *PRCreator) generatePRBodyWithIssue(taskMsg *TaskMessage, result *AITaskResponse, issueNumber string) string {
	body := fmt.Sprintf("## Task Description\n%s\n\n", taskMsg.Tasks)

	// Add issue link if provided
	if issueNumber != "" {
		body += fmt.Sprintf("Closes #%s\n\n", issueNumber)
	}

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