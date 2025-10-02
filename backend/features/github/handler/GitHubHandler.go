package handler

import (
	"context"
	"fmt"
	"log"
	"main/features/github/model/response"
	"main/features/github/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type GitHubHandler struct {
	githubService *service.GitHubService
}

func NewGitHubHandler() *GitHubHandler {
	return &GitHubHandler{
		githubService: service.NewGitHubService(),
	}
}

// @Summary Sync user repositories
// @Description Clone new repositories and update existing ones for authenticated user
// @Tags github
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /api/v1/github/sync-repositories [post]
func (h *GitHubHandler) SyncRepositories(c echo.Context) error {
	ctx := c.Request().Context()

	// Get access token from header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Authorization header required",
		})
	}

	// Extract token (remove "Bearer " prefix if present)
	accessToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	// Get target directory from environment or use default
	targetDir := os.Getenv("REPOSITORIES_DIR")
	if targetDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Failed to get home directory: %v", err)
			targetDir = "/tmp/repositories"
		} else {
			targetDir = filepath.Join(homeDir, "project", "git-repository")
		}
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("Failed to create target directory %s: %v", targetDir, err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create repositories directory",
		})
	}

	// Clone or update repositories
	if err := h.githubService.CloneOrUpdateRepositories(ctx, accessToken, targetDir); err != nil {
		log.Printf("Failed to sync repositories: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to sync repositories",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Repositories synced successfully",
		"target_directory": targetDir,
	})
}

// @Summary Get user repositories
// @Description Get list of user repositories from GitHub
// @Tags github
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "repositories list"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /api/v1/github/repositories [get]
func (h *GitHubHandler) GetRepositories(c echo.Context) error {
	ctx := c.Request().Context()

	// Get access token from header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Authorization header required",
		})
	}

	// Extract token (remove "Bearer " prefix if present)
	accessToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	// Get repositories
	repos, err := h.githubService.GetUserRepositories(ctx, accessToken)
	if err != nil {
		log.Printf("Failed to get repositories: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get repositories",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"repositories": repos,
		"count": len(repos),
	})
}

// @Summary Get authenticated user info
// @Description Get current authenticated user information from GitHub
// @Tags github
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "user info"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /api/v1/github/user [get]
func (h *GitHubHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()

	// Get access token from header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Authorization header required",
		})
	}

	// Extract token (remove "Bearer " prefix if present)
	accessToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	// Get user info
	user, err := h.githubService.GetAuthenticatedUser(ctx, accessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get user info",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// SyncRepositoriesOnLogin is called during login to automatically sync repositories
func (h *GitHubHandler) SyncRepositoriesOnLogin(ctx context.Context, accessToken string) error {
	// Get target directory from environment or use default
	targetDir := os.Getenv("REPOSITORIES_DIR")
	if targetDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Failed to get home directory: %v", err)
			targetDir = "/tmp/repositories"
		} else {
			targetDir = filepath.Join(homeDir, "project", "git-repository")
		}
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("Failed to create target directory %s: %v", targetDir, err)
		return err
	}

	// Clone or update repositories in background
	go func() {
		if err := h.githubService.CloneOrUpdateRepositories(context.Background(), accessToken, targetDir); err != nil {
			log.Printf("Background repository sync failed: %v", err)
		} else {
			log.Printf("Background repository sync completed successfully")
		}
	}()

	return nil
}

// @Summary Merge pull request
// @Description Merge a GitHub pull request
// @Tags github
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Param number path int true "Pull request number"
// @Param request body response.MergePullRequestRequest true "Merge request"
// @Success 200 {object} response.MergePullRequestResponse
// @Failure 403 {object} map[string]interface{} "permission denied"
// @Failure 404 {object} map[string]interface{} "not found"
// @Failure 422 {object} map[string]interface{} "not mergeable"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge [put]
func (h *GitHubHandler) MergePullRequest(c echo.Context) error {
	ctx := c.Request().Context()

	// Get path parameters
	owner := c.Param("owner")
	repo := c.Param("repo")
	prNumber := c.Param("number")

	if owner == "" || repo == "" || prNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "owner, repo, and number parameters are required",
		})
	}

	// Convert PR number to int
	var prNum int
	if _, err := fmt.Sscanf(prNumber, "%d", &prNum); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "invalid pull request number",
		})
	}

	// Get GitHub token from environment (NOT from request for security)
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Printf("GITHUB_TOKEN environment variable not set")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "GitHub token not configured",
		})
	}

	// Bind request body
	var req response.MergePullRequestRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate merge method
	if req.MergeMethod != "" && req.MergeMethod != "merge" && req.MergeMethod != "squash" && req.MergeMethod != "rebase" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "merge_method must be one of: merge, squash, rebase",
		})
	}

	// Merge pull request
	result, err := h.githubService.MergePullRequest(ctx, githubToken, owner, repo, prNum, &req)
	if err != nil {
		log.Printf("Failed to merge PR #%d in %s/%s: %v", prNum, owner, repo, err)

		// Check error type and return appropriate status code
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Pull request not found",
			})
		}
		if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "authentication failed") {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error": "Permission denied",
			})
		}
		if strings.Contains(errMsg, "validation failed") {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"error": "Pull request not mergeable",
				"details": errMsg,
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to merge pull request",
			"details": errMsg,
		})
	}

	return c.JSON(http.StatusOK, result)
}