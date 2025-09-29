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
	"regexp"
	"strings"
	"time"
)

type GitHubService struct {
	httpClient *http.Client
	BaseURL    string
}

func NewGitHubService() *GitHubService {
	return &GitHubService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: "https://api.github.com",
	}
}

// Issue Management

// CreateIssue creates a new GitHub issue
func (s *GitHubService) CreateIssue(ctx context.Context, accessToken, owner, repo string, req *response.CreateIssueRequest) (*response.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", s.BaseURL, owner, repo)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
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
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", s.BaseURL, owner, repo, issueNumber)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue update request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
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
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", s.BaseURL, owner, repo)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pull request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
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
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/requested_reviewers", s.BaseURL, owner, repo, prNumber)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal reviewers request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
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
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", s.BaseURL, owner, repo, prNumber)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

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
	return fmt.Sprintf(`<!-- AUTO-GENERATED -->
## Task Description
%s

**Repository:** %s
**Working Directory:** %s
**Request ID:** %s
**Created:** %s

## Implementation Status
- [ ] Analysis complete
- [ ] Implementation complete
- [ ] Testing complete
- [ ] Documentation updated

## Execution Context
This issue tracks the automated execution of a task through the local-backend system. The task will be executed in an isolated branch and automatically create a pull request upon completion.

---
*This issue was automatically created by the task execution system.*`,
		taskDescription,
		repositoryName,
		workingDir,
		requestID,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// GeneratePRTemplate generates a standardized PR template for tasks
func (s *GitHubService) GeneratePRTemplate(taskDescription, repositoryName, branchName, issueNumber, executionTime string) string {
	return fmt.Sprintf(`<!-- AUTO-GENERATED -->
## Summary
%s

**Repository:** %s
**Branch:** %s
**Execution Time:** %s

Closes #%s

## Changes Made
This pull request contains the automated implementation of the task described in the linked issue. All changes were generated through the Claude Code task execution system.

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Code review completed

## Implementation Details
- **Automated Execution:** ✅
- **Branch Isolation:** ✅
- **Task Completion:** ✅

---
*This PR was automatically created by the task execution system.*`,
		taskDescription,
		repositoryName,
		branchName,
		executionTime,
		issueNumber,
	)
}

// Utility Functions

// ExtractIssueNumber extracts issue number from GitHub issue URL
func (s *GitHubService) ExtractIssueNumber(issueURL string) string {
	re := regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/issues/(\d+)`)
	matches := re.FindStringSubmatch(issueURL)
	if len(matches) < 2 {
		re = regexp.MustCompile(`https://api\.github\.com/repos/[^/]+/[^/]+/issues/(\d+)`)
		matches = re.FindStringSubmatch(issueURL)
		if len(matches) < 2 {
			return ""
		}
	}
	return matches[1]
}

// ExtractPRNumber extracts PR number from GitHub pull request URL
func (s *GitHubService) ExtractPRNumber(prURL string) string {
	re := regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/pull/(\d+)`)
	matches := re.FindStringSubmatch(prURL)
	if len(matches) < 2 {
		re = regexp.MustCompile(`https://api\.github\.com/repos/[^/]+/[^/]+/pulls/(\d+)`)
		matches = re.FindStringSubmatch(prURL)
		if len(matches) < 2 {
			return ""
		}
	}
	return matches[1]
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