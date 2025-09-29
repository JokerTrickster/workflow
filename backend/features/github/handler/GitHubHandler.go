package handler

import (
	"context"
	"log"
	"main/features/github/service"
	"net/http"
	"os"
	"path/filepath"

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