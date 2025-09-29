package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/features/github/model/response"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GitHubService struct {
	httpClient *http.Client
}

func NewGitHubService() *GitHubService {
	return &GitHubService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *GitHubService) GetAuthenticatedUser(ctx context.Context, accessToken string) (*response.User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var user response.User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (s *GitHubService) GetUserRepositories(ctx context.Context, accessToken string) ([]response.Repository, error) {
	var allRepos []response.Repository
	page := 1
	perPage := 100

	for {
		repos, hasMore, err := s.getRepositoriesPage(ctx, accessToken, page, perPage)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if !hasMore {
			break
		}
		page++
	}

	log.Printf("Retrieved %d repositories from GitHub", len(allRepos))
	return allRepos, nil
}

func (s *GitHubService) getRepositoriesPage(ctx context.Context, accessToken string, page, perPage int) ([]response.Repository, bool, error) {
	url := fmt.Sprintf("https://api.github.com/user/repos?page=%d&per_page=%d&sort=updated&direction=desc", page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response: %w", err)
	}

	var repos []response.Repository
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal repositories: %w", err)
	}

	// Check if there are more pages
	hasMore := len(repos) == perPage

	return repos, hasMore, nil
}

func (s *GitHubService) CloneOrUpdateRepositories(ctx context.Context, accessToken string, targetDir string) error {
	// Get user info first
	user, err := s.GetAuthenticatedUser(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Get all repositories
	repos, err := s.GetUserRepositories(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get repositories: %w", err)
	}

	// Create user directory
	userDir := filepath.Join(targetDir, user.Login)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	log.Printf("Processing %d repositories for user %s", len(repos), user.Login)

	// Process repositories concurrently with limit
	const maxConcurrent = 5
	semaphore := make(chan struct{}, maxConcurrent)
	errChan := make(chan error, len(repos))

	for _, repo := range repos {
		go func(repo response.Repository) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			err := s.cloneOrUpdateRepository(ctx, repo, userDir, accessToken)
			errChan <- err
		}(repo)
	}

	// Collect results
	var errors []string
	for i := 0; i < len(repos); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err.Error())
			log.Printf("Repository operation failed: %v", err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed operations: %s", strings.Join(errors, "; "))
	}

	log.Printf("Successfully processed all %d repositories", len(repos))
	return nil
}

func (s *GitHubService) cloneOrUpdateRepository(ctx context.Context, repo response.Repository, userDir, accessToken string) error {
	repoPath := filepath.Join(userDir, repo.Name)

	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Repository exists, update it
		return s.updateRepository(ctx, repoPath, repo.Name)
	} else if os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		return s.cloneRepository(ctx, repo, userDir, accessToken)
	} else {
		return fmt.Errorf("failed to check repository status for %s: %w", repo.Name, err)
	}
}

func (s *GitHubService) cloneRepository(ctx context.Context, repo response.Repository, userDir, accessToken string) error {
	log.Printf("Cloning repository: %s", repo.Name)

	// Use HTTPS clone URL with token authentication
	cloneURL := fmt.Sprintf("https://%s@github.com/%s.git", accessToken, repo.FullName)

	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, filepath.Join(userDir, repo.Name))
	cmd.Dir = userDir

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone %s: %w, output: %s", repo.Name, err, string(output))
	}

	log.Printf("Successfully cloned repository: %s", repo.Name)
	return nil
}

func (s *GitHubService) updateRepository(ctx context.Context, repoPath, repoName string) error {
	log.Printf("Updating repository: %s", repoName)

	// Change to repository directory and pull latest changes
	cmd := exec.CommandContext(ctx, "git", "pull", "origin")
	cmd.Dir = repoPath

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try to reset to origin/main if pull fails
		resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard", "origin/main")
		resetCmd.Dir = repoPath
		if resetErr := resetCmd.Run(); resetErr != nil {
			return fmt.Errorf("failed to pull and reset %s: pull error: %w, output: %s, reset error: %v",
				repoName, err, string(output), resetErr)
		}
		log.Printf("Reset repository %s to origin/main due to pull failure", repoName)
	}

	log.Printf("Successfully updated repository: %s", repoName)
	return nil
}

// Issue Management

// CreateIssue creates a new GitHub issue
func (s *GitHubService) CreateIssue(ctx context.Context, accessToken, owner, repo string, req *response.CreateIssueRequest) (*response.Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, s.handleGitHubAPIError(resp.StatusCode, body, fmt.Sprintf("create issue in %s/%s", owner, repo))
	}

	var issue response.Issue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issue: %w", err)
	}

	log.Printf("Successfully created issue #%d in %s/%s: %s", issue.Number, owner, repo, issue.Title)
	return &issue, nil
}

// UpdateIssue updates an existing GitHub issue
func (s *GitHubService) UpdateIssue(ctx context.Context, accessToken, owner, repo string, issueNumber int, req *response.UpdateIssueRequest) (*response.Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, issueNumber)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue update request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, s.handleGitHubAPIError(resp.StatusCode, body, fmt.Sprintf("update issue #%d in %s/%s", issueNumber, owner, repo))
	}

	var issue response.Issue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issue: %w", err)
	}

	log.Printf("Successfully updated issue #%d in %s/%s", issue.Number, owner, repo)
	return &issue, nil
}

// CloseIssue closes a GitHub issue
func (s *GitHubService) CloseIssue(ctx context.Context, accessToken, owner, repo string, issueNumber int) (*response.Issue, error) {
	state := "closed"
	req := &response.UpdateIssueRequest{
		State: &state,
	}
	return s.UpdateIssue(ctx, accessToken, owner, repo, issueNumber, req)
}

// Pull Request Management

// CreatePullRequest creates a new GitHub pull request
func (s *GitHubService) CreatePullRequest(ctx context.Context, accessToken, owner, repo string, req *response.CreatePullRequestRequest) (*response.PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pull request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, s.handleGitHubAPIError(resp.StatusCode, body, fmt.Sprintf("create pull request in %s/%s", owner, repo))
	}

	var pr response.PullRequest
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pull request: %w", err)
	}

	log.Printf("Successfully created pull request #%d in %s/%s: %s", pr.Number, owner, repo, pr.Title)
	return &pr, nil
}

// AssignReviewers adds reviewers to a pull request
func (s *GitHubService) AssignReviewers(ctx context.Context, accessToken, owner, repo string, prNumber int, req *response.RequestReviewersRequest) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, prNumber)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal reviewers request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to assign reviewers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return s.handleGitHubAPIError(resp.StatusCode, body, fmt.Sprintf("assign reviewers to PR #%d in %s/%s", prNumber, owner, repo))
	}

	log.Printf("Successfully assigned reviewers to PR #%d in %s/%s", prNumber, owner, repo)
	return nil
}

// GetPullRequestStatus gets the status of a pull request
func (s *GitHubService) GetPullRequestStatus(ctx context.Context, accessToken, owner, repo string, prNumber int) (*response.PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, s.handleGitHubAPIError(resp.StatusCode, body, fmt.Sprintf("get PR #%d status in %s/%s", prNumber, owner, repo))
	}

	var pr response.PullRequest
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pull request: %w", err)
	}

	return &pr, nil
}

// Template Generation

// GenerateIssueTemplate generates a standardized issue template for tasks
func (s *GitHubService) GenerateIssueTemplate(taskDescription, repositoryName, workingDir, requestID string) string {
	return fmt.Sprintf(`# Task: %s

**Repository**: %s
**Executed by**: Local Backend Task System
**Created**: %s

## Task Details
%s

## Execution Context
- **Branch**: Will be created automatically
- **Working Directory**: %s
- **Task ID**: %s

---
*This issue was automatically created by the task execution system.*`,
		taskDescription,
		repositoryName,
		time.Now().Format("2006-01-02 15:04:05"),
		taskDescription,
		workingDir,
		requestID,
	)
}

// GeneratePRTemplate generates a standardized PR template for tasks
func (s *GitHubService) GeneratePRTemplate(taskDescription, repositoryName, branchName, issueNumber, executionTime string) string {
	return fmt.Sprintf(`# Task Implementation: %s

**Closes**: #%s

## Changes Made
Automated implementation via Claude Code task execution

## Task Context
- **Original Task**: %s
- **Repository**: %s
- **Branch**: %s
- **Execution Time**: %s

## Review Checklist
- [ ] Code follows project conventions
- [ ] Changes are well-tested
- [ ] Documentation updated if needed
- [ ] No breaking changes introduced

---
*This PR was automatically created by the task execution system.*`,
		taskDescription,
		issueNumber,
		taskDescription,
		repositoryName,
		branchName,
		executionTime,
	)
}

// Utility Functions

// ExtractIssueNumber extracts issue number from GitHub issue URL
func (s *GitHubService) ExtractIssueNumber(issueURL string) (int, error) {
	re := regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/issues/(\d+)`)
	matches := re.FindStringSubmatch(issueURL)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid issue URL format: %s", issueURL)
	}

	issueNumber, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid issue number in URL: %s", issueURL)
	}

	return issueNumber, nil
}

// ExtractPRNumber extracts PR number from GitHub pull request URL
func (s *GitHubService) ExtractPRNumber(prURL string) (int, error) {
	re := regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/pull/(\d+)`)
	matches := re.FindStringSubmatch(prURL)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid PR URL format: %s", prURL)
	}

	prNumber, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid PR number in URL: %s", prURL)
	}

	return prNumber, nil
}

// handleGitHubAPIError handles GitHub API errors with appropriate error messages
func (s *GitHubService) handleGitHubAPIError(statusCode int, body []byte, operation string) error {
	switch statusCode {
	case http.StatusNotFound:
		return fmt.Errorf("GitHub resource not found for operation: %s", operation)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub API authentication failed for operation: %s", operation)
	case http.StatusForbidden:
		// Check if it's a rate limit error
		var githubError struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
		}
		if err := json.Unmarshal(body, &githubError); err == nil {
			if strings.Contains(githubError.Message, "rate limit") {
				return fmt.Errorf("GitHub API rate limit exceeded for operation: %s", operation)
			}
			return fmt.Errorf("GitHub API access forbidden for operation %s: %s", operation, githubError.Message)
		}
		return fmt.Errorf("GitHub API access forbidden for operation: %s", operation)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("GitHub API validation failed for operation: %s, response: %s", operation, string(body))
	case http.StatusTooManyRequests:
		return fmt.Errorf("GitHub API rate limit exceeded for operation: %s", operation)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("GitHub API is temporarily unavailable for operation: %s", operation)
	default:
		return fmt.Errorf("GitHub API returned status %d for operation %s: %s", statusCode, operation, string(body))
	}
}