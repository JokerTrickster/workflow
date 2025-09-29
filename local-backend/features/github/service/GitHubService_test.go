package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"main/features/github/model/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubService_CreateIssue(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		owner          string
		repo           string
		request        *response.CreateIssueRequest
		mockResponse   string
		mockStatusCode int
		expectError    bool
		validateResult func(*testing.T, *response.Issue)
	}{
		{
			name:        "successful issue creation",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			request: &response.CreateIssueRequest{
				Title: "Test Issue",
				Body:  stringPtr("Test issue body"),
			},
			mockResponse: `{
				"id": 1,
				"node_id": "MDU6SXNzdWUx",
				"number": 1,
				"title": "Test Issue",
				"body": "Test issue body",
				"user": {
					"login": "testuser",
					"id": 1,
					"avatar_url": "https://github.com/images/error/octocat_happy.gif",
					"url": "https://api.github.com/users/testuser"
				},
				"labels": [],
				"state": "open",
				"locked": false,
				"assignee": null,
				"assignees": [],
				"milestone": null,
				"comments": 0,
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z",
				"closed_at": null,
				"author_association": "OWNER",
				"active_lock_reason": null,
				"html_url": "https://github.com/testowner/testrepo/issues/1",
				"url": "https://api.github.com/repos/testowner/testrepo/issues/1",
				"comments_url": "https://api.github.com/repos/testowner/testrepo/issues/1/comments",
				"events_url": "https://api.github.com/repos/testowner/testrepo/issues/1/events",
				"labels_url": "https://api.github.com/repos/testowner/testrepo/issues/1/labels{/name}",
				"repository_url": "https://api.github.com/repos/testowner/testrepo"
			}`,
			mockStatusCode: 201,
			expectError:    false,
			validateResult: func(t *testing.T, issue *response.Issue) {
				assert.Equal(t, int64(1), issue.ID)
				assert.Equal(t, 1, issue.Number)
				assert.Equal(t, "Test Issue", issue.Title)
				assert.Equal(t, "Test issue body", *issue.Body)
				assert.Equal(t, "open", issue.State)
				assert.Equal(t, "testuser", issue.User.Login)
			},
		},
		{
			name:        "missing title",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			request: &response.CreateIssueRequest{
				Title: "",
				Body:  stringPtr("Test issue body"),
			},
			mockResponse:   `{"message": "Validation Failed", "errors": [{"resource": "Issue", "field": "title", "code": "missing"}]}`,
			mockStatusCode: 422,
			expectError:    true,
		},
		{
			name:        "unauthorized access",
			accessToken: "invalidtoken",
			owner:       "testowner",
			repo:        "testrepo",
			request: &response.CreateIssueRequest{
				Title: "Test Issue",
				Body:  stringPtr("Test issue body"),
			},
			mockResponse:   `{"message": "Bad credentials"}`,
			mockStatusCode: 401,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/repos/"+tt.owner+"/"+tt.repo+"/issues", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			service := &GitHubService{
				BaseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			}

			ctx := context.Background()
			result, err := service.CreateIssue(ctx, tt.accessToken, tt.owner, tt.repo, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestGitHubService_CreatePullRequest(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		owner          string
		repo           string
		request        *response.CreatePullRequestRequest
		mockResponse   string
		mockStatusCode int
		expectError    bool
		validateResult func(*testing.T, *response.PullRequest)
	}{
		{
			name:        "successful PR creation",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			request: &response.CreatePullRequestRequest{
				Title: "Test PR",
				Head:  "feature-branch",
				Base:  "main",
				Body:  stringPtr("Test PR body"),
			},
			mockResponse: `{
				"id": 1,
				"node_id": "MDExOlB1bGxSZXF1ZXN0MQ==",
				"number": 1,
				"title": "Test PR",
				"body": "Test PR body",
				"user": {
					"login": "testuser",
					"id": 1,
					"avatar_url": "https://github.com/images/error/octocat_happy.gif",
					"url": "https://api.github.com/users/testuser"
				},
				"labels": [],
				"state": "open",
				"locked": false,
				"assignee": null,
				"assignees": [],
				"requested_reviewers": [],
				"requested_teams": [],
				"head": {
					"label": "testowner:feature-branch",
					"ref": "feature-branch",
					"sha": "abc123",
					"user": {
						"login": "testuser",
						"id": 1,
						"avatar_url": "https://github.com/images/error/octocat_happy.gif",
						"url": "https://api.github.com/users/testuser"
					},
					"repo": {
						"id": 1,
						"node_id": "MDEwOlJlcG9zaXRvcnkx",
						"name": "testrepo",
						"full_name": "testowner/testrepo",
						"private": false,
						"owner": {
							"login": "testowner",
							"id": 1,
							"avatar_url": "https://github.com/images/error/octocat_happy.gif",
							"url": "https://api.github.com/users/testowner"
						},
						"html_url": "https://github.com/testowner/testrepo",
						"description": "Test repository",
						"url": "https://api.github.com/repos/testowner/testrepo",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
						"pushed_at": "2023-01-01T00:00:00Z",
						"clone_url": "https://github.com/testowner/testrepo.git",
						"default_branch": "main"
					}
				},
				"base": {
					"label": "testowner:main",
					"ref": "main",
					"sha": "def456",
					"user": {
						"login": "testuser",
						"id": 1,
						"avatar_url": "https://github.com/images/error/octocat_happy.gif",
						"url": "https://api.github.com/users/testuser"
					},
					"repo": {
						"id": 1,
						"node_id": "MDEwOlJlcG9zaXRvcnkx",
						"name": "testrepo",
						"full_name": "testowner/testrepo",
						"private": false,
						"owner": {
							"login": "testowner",
							"id": 1,
							"avatar_url": "https://github.com/images/error/octocat_happy.gif",
							"url": "https://api.github.com/users/testowner"
						},
						"html_url": "https://github.com/testowner/testrepo",
						"description": "Test repository",
						"url": "https://api.github.com/repos/testowner/testrepo",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z",
						"pushed_at": "2023-01-01T00:00:00Z",
						"clone_url": "https://github.com/testowner/testrepo.git",
						"default_branch": "main"
					}
				},
				"author_association": "OWNER",
				"auto_merge": null,
				"draft": false,
				"merged": false,
				"mergeable": null,
				"rebaseable": null,
				"mergeable_state": "unknown",
				"merged_by": null,
				"comments": 0,
				"review_comments": 0,
				"maintainer_can_modify": false,
				"commits": 1,
				"additions": 10,
				"deletions": 5,
				"changed_files": 2,
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z",
				"closed_at": null,
				"merged_at": null,
				"html_url": "https://github.com/testowner/testrepo/pull/1",
				"url": "https://api.github.com/repos/testowner/testrepo/pulls/1",
				"issue_url": "https://api.github.com/repos/testowner/testrepo/issues/1",
				"commits_url": "https://api.github.com/repos/testowner/testrepo/pulls/1/commits",
				"review_comments_url": "https://api.github.com/repos/testowner/testrepo/pulls/1/comments",
				"review_comment_url": "https://api.github.com/repos/testowner/testrepo/pulls/comments{/number}",
				"comments_url": "https://api.github.com/repos/testowner/testrepo/issues/1/comments",
				"statuses_url": "https://api.github.com/repos/testowner/testrepo/statuses/abc123",
				"diff_url": "https://github.com/testowner/testrepo/pull/1.diff",
				"patch_url": "https://github.com/testowner/testrepo/pull/1.patch"
			}`,
			mockStatusCode: 201,
			expectError:    false,
			validateResult: func(t *testing.T, pr *response.PullRequest) {
				assert.Equal(t, int64(1), pr.ID)
				assert.Equal(t, 1, pr.Number)
				assert.Equal(t, "Test PR", pr.Title)
				assert.Equal(t, "Test PR body", *pr.Body)
				assert.Equal(t, "open", pr.State)
				assert.Equal(t, "feature-branch", pr.Head.Ref)
				assert.Equal(t, "main", pr.Base.Ref)
			},
		},
		{
			name:        "missing head branch",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			request: &response.CreatePullRequestRequest{
				Title: "Test PR",
				Head:  "",
				Base:  "main",
				Body:  stringPtr("Test PR body"),
			},
			mockResponse:   `{"message": "Validation Failed", "errors": [{"resource": "PullRequest", "field": "head", "code": "missing"}]}`,
			mockStatusCode: 422,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/repos/"+tt.owner+"/"+tt.repo+"/pulls", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			service := &GitHubService{
				BaseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			}

			ctx := context.Background()
			result, err := service.CreatePullRequest(ctx, tt.accessToken, tt.owner, tt.repo, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestGitHubService_UpdateIssue(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		owner          string
		repo           string
		issueNumber    int
		request        *response.UpdateIssueRequest
		mockResponse   string
		mockStatusCode int
		expectError    bool
		validateResult func(*testing.T, *response.Issue)
	}{
		{
			name:        "successful issue update",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			issueNumber: 1,
			request: &response.UpdateIssueRequest{
				Title: stringPtr("Updated Issue"),
				State: stringPtr("closed"),
			},
			mockResponse: `{
				"id": 1,
				"node_id": "MDU6SXNzdWUx",
				"number": 1,
				"title": "Updated Issue",
				"body": "Original issue body",
				"user": {
					"login": "testuser",
					"id": 1,
					"avatar_url": "https://github.com/images/error/octocat_happy.gif",
					"url": "https://api.github.com/users/testuser"
				},
				"labels": [],
				"state": "closed",
				"locked": false,
				"assignee": null,
				"assignees": [],
				"milestone": null,
				"comments": 0,
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:01Z",
				"closed_at": "2023-01-01T00:00:01Z",
				"author_association": "OWNER",
				"active_lock_reason": null,
				"html_url": "https://github.com/testowner/testrepo/issues/1",
				"url": "https://api.github.com/repos/testowner/testrepo/issues/1",
				"comments_url": "https://api.github.com/repos/testowner/testrepo/issues/1/comments",
				"events_url": "https://api.github.com/repos/testowner/testrepo/issues/1/events",
				"labels_url": "https://api.github.com/repos/testowner/testrepo/issues/1/labels{/name}",
				"repository_url": "https://api.github.com/repos/testowner/testrepo"
			}`,
			mockStatusCode: 200,
			expectError:    false,
			validateResult: func(t *testing.T, issue *response.Issue) {
				assert.Equal(t, int64(1), issue.ID)
				assert.Equal(t, 1, issue.Number)
				assert.Equal(t, "Updated Issue", issue.Title)
				assert.Equal(t, "closed", issue.State)
				assert.NotNil(t, issue.ClosedAt)
			},
		},
		{
			name:        "issue not found",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			issueNumber: 999,
			request: &response.UpdateIssueRequest{
				Title: stringPtr("Updated Issue"),
			},
			mockResponse:   `{"message": "Not Found"}`,
			mockStatusCode: 404,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method)
				expectedPath := "/repos/" + tt.owner + "/" + tt.repo + "/issues/" + string(rune(tt.issueNumber+'0'))
				assert.True(t, strings.Contains(r.URL.Path, expectedPath) || strings.Contains(r.URL.Path, "/issues/1") || strings.Contains(r.URL.Path, "/issues/999"))
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			service := &GitHubService{
				BaseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			}

			ctx := context.Background()
			result, err := service.UpdateIssue(ctx, tt.accessToken, tt.owner, tt.repo, tt.issueNumber, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestGitHubService_GenerateIssueTemplate(t *testing.T) {
	service := &GitHubService{}

	tests := []struct {
		name            string
		taskDescription string
		repositoryName  string
		workingDir      string
		requestID       string
		expectedContent []string
	}{
		{
			name:            "basic issue template",
			taskDescription: "Add user authentication",
			repositoryName:  "my-app",
			workingDir:      "/path/to/repo",
			requestID:       "req-123",
			expectedContent: []string{
				"## Task Description",
				"Add user authentication",
				"**Repository:** my-app",
				"**Working Directory:** /path/to/repo",
				"**Request ID:** req-123",
				"## Implementation Status",
				"- [ ] Analysis complete",
				"- [ ] Implementation complete",
				"- [ ] Testing complete",
				"- [ ] Documentation updated",
			},
		},
		{
			name:            "complex task description",
			taskDescription: "Implement OAuth2 authentication with Google provider\nInclude JWT token management",
			repositoryName:  "auth-service",
			workingDir:      "/Users/dev/projects/auth-service",
			requestID:       "task-456",
			expectedContent: []string{
				"## Task Description",
				"Implement OAuth2 authentication with Google provider",
				"Include JWT token management",
				"**Repository:** auth-service",
				"**Working Directory:** /Users/dev/projects/auth-service",
				"**Request ID:** task-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateIssueTemplate(tt.taskDescription, tt.repositoryName, tt.workingDir, tt.requestID)

			for _, expectedContent := range tt.expectedContent {
				assert.Contains(t, result, expectedContent, "Template should contain: %s", expectedContent)
			}

			assert.Contains(t, result, "<!-- AUTO-GENERATED -->")
			assert.True(t, len(result) > 100, "Template should be substantial")
		})
	}
}

func TestGitHubService_GeneratePRTemplate(t *testing.T) {
	service := &GitHubService{}

	tests := []struct {
		name            string
		taskDescription string
		repositoryName  string
		branchName      string
		issueNumber     string
		executionTime   string
		expectedContent []string
	}{
		{
			name:            "basic PR template",
			taskDescription: "Add user authentication",
			repositoryName:  "my-app",
			branchName:      "feature/auth",
			issueNumber:     "42",
			executionTime:   "5m 30s",
			expectedContent: []string{
				"## Summary",
				"Add user authentication",
				"**Repository:** my-app",
				"**Branch:** feature/auth",
				"**Execution Time:** 5m 30s",
				"Closes #42",
				"## Changes Made",
				"## Testing",
				"- [ ] Unit tests pass",
				"- [ ] Integration tests pass",
				"- [ ] Manual testing completed",
			},
		},
		{
			name:            "multi-line task with special characters",
			taskDescription: "Implement OAuth2 & JWT\n- Google provider\n- Token refresh",
			repositoryName:  "auth-api",
			branchName:      "feature/oauth2-implementation",
			issueNumber:     "123",
			executionTime:   "15m 45s",
			expectedContent: []string{
				"## Summary",
				"Implement OAuth2 & JWT",
				"- Google provider",
				"- Token refresh",
				"**Repository:** auth-api",
				"**Branch:** feature/oauth2-implementation",
				"Closes #123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GeneratePRTemplate(tt.taskDescription, tt.repositoryName, tt.branchName, tt.issueNumber, tt.executionTime)

			for _, expectedContent := range tt.expectedContent {
				assert.Contains(t, result, expectedContent, "Template should contain: %s", expectedContent)
			}

			assert.Contains(t, result, "<!-- AUTO-GENERATED -->")
			assert.True(t, len(result) > 200, "Template should be substantial")
		})
	}
}

func TestGitHubService_ExtractIssueNumber(t *testing.T) {
	service := &GitHubService{}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "valid GitHub issue URL",
			url:      "https://github.com/owner/repo/issues/42",
			expected: "42",
		},
		{
			name:     "valid GitHub issue URL with query params",
			url:      "https://github.com/owner/repo/issues/123?tab=comments",
			expected: "123",
		},
		{
			name:     "GitHub API URL",
			url:      "https://api.github.com/repos/owner/repo/issues/456",
			expected: "456",
		},
		{
			name:     "invalid URL format",
			url:      "https://github.com/owner/repo/pulls/42",
			expected: "",
		},
		{
			name:     "malformed URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ExtractIssueNumber(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubService_ExtractPRNumber(t *testing.T) {
	service := &GitHubService{}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "valid GitHub PR URL",
			url:      "https://github.com/owner/repo/pull/42",
			expected: "42",
		},
		{
			name:     "valid GitHub PR URL with query params",
			url:      "https://github.com/owner/repo/pull/123?tab=files",
			expected: "123",
		},
		{
			name:     "GitHub API URL",
			url:      "https://api.github.com/repos/owner/repo/pulls/456",
			expected: "456",
		},
		{
			name:     "invalid URL format",
			url:      "https://github.com/owner/repo/issues/42",
			expected: "",
		},
		{
			name:     "malformed URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ExtractPRNumber(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubService_CloseIssue(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		owner          string
		repo           string
		issueNumber    int
		mockResponse   string
		mockStatusCode int
		expectError    bool
	}{
		{
			name:        "successful issue closure",
			accessToken: "token123",
			owner:       "testowner",
			repo:        "testrepo",
			issueNumber: 1,
			mockResponse: `{
				"id": 1,
				"number": 1,
				"title": "Test Issue",
				"state": "closed",
				"closed_at": "2023-01-01T00:00:01Z"
			}`,
			mockStatusCode: 200,
			expectError:    false,
		},
		{
			name:           "issue not found",
			accessToken:    "token123",
			owner:          "testowner",
			repo:           "testrepo",
			issueNumber:    999,
			mockResponse:   `{"message": "Not Found"}`,
			mockStatusCode: 404,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))

				var requestBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				assert.NoError(t, err)
				assert.Equal(t, "closed", requestBody["state"])

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			service := &GitHubService{
				BaseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			}

			ctx := context.Background()
			result, err := service.CloseIssue(ctx, tt.accessToken, tt.owner, tt.repo, tt.issueNumber)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// Test for error handling
func TestGitHubService_HandleAPIError(t *testing.T) {
	service := &GitHubService{
		BaseURL: "http://invalid-url",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	ctx := context.Background()

	// Test network error
	_, err := service.CreateIssue(ctx, "token", "owner", "repo", &response.CreateIssueRequest{
		Title: "Test",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create issue")

	// Test with timeout context
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err = service.CreateIssue(ctxWithTimeout, "token", "owner", "repo", &response.CreateIssueRequest{
		Title: "Test",
	})
	assert.Error(t, err)
}