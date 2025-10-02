package handler

import (
	"github.com/labstack/echo/v4"
)

func InitGitHubHandler(e *echo.Echo) *GitHubHandler {
	githubHandler := NewGitHubHandler()

	// GitHub API 라우트 그룹
	githubGroup := e.Group("/api/v1/github")

	// GitHub API 엔드포인트들
	githubGroup.GET("/user", githubHandler.GetUser)
	githubGroup.GET("/repositories", githubHandler.GetRepositories)
	githubGroup.POST("/sync-repositories", githubHandler.SyncRepositories)
	githubGroup.PUT("/repos/:owner/:repo/pulls/:number/merge", githubHandler.MergePullRequest)

	return githubHandler
}