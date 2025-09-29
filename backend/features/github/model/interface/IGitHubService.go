package _interface

import (
	"context"
	"main/features/github/model/response"
)

type IGitHubService interface {
	// GetUserRepositories gets all repositories for the authenticated user
	GetUserRepositories(ctx context.Context, accessToken string) ([]response.Repository, error)

	// CloneOrUpdateRepositories clones new repositories and updates existing ones
	CloneOrUpdateRepositories(ctx context.Context, accessToken string, targetDir string) error

	// GetAuthenticatedUser gets the current authenticated user info
	GetAuthenticatedUser(ctx context.Context, accessToken string) (*response.User, error)
}