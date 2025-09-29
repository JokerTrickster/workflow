package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWorkflowHistoriesModel(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the WorkflowHistories model
	err = db.AutoMigrate(&WorkflowHistories{})
	assert.NoError(t, err)

	// Test basic CRUD operations
	t.Run("Create WorkflowHistories with Git fields", func(t *testing.T) {
		now := time.Now()
		githubIssueURL := "https://github.com/owner/repo/issues/123"
		githubPRURL := "https://github.com/owner/repo/pull/456"
		branchName := "feature/test-branch"
		cleanupStatus := "pending"

		workflow := WorkflowHistories{
			RequestID:        "test-request-123",
			Status:           "processing",
			Tasks:            "Test task description",
			RepositoryName:   "test-repo",
			Provider:         "claude",
			CreatedAt:        now,
			GitHubIssueURL:   &githubIssueURL,
			GitHubPRURL:      &githubPRURL,
			BranchName:       &branchName,
			CleanupStatus:    &cleanupStatus,
		}

		result := db.Create(&workflow)
		assert.NoError(t, result.Error)
		assert.NotZero(t, workflow.ID)
	})

	t.Run("Read WorkflowHistories with Git fields", func(t *testing.T) {
		var workflow WorkflowHistories
		result := db.Where("request_id = ?", "test-request-123").First(&workflow)
		assert.NoError(t, result.Error)

		assert.Equal(t, "test-request-123", workflow.RequestID)
		assert.Equal(t, "processing", workflow.Status)
		assert.Equal(t, "Test task description", workflow.Tasks)
		assert.Equal(t, "test-repo", workflow.RepositoryName)
		assert.Equal(t, "claude", workflow.Provider)

		// Test Git workflow fields
		assert.NotNil(t, workflow.GitHubIssueURL)
		assert.Equal(t, "https://github.com/owner/repo/issues/123", *workflow.GitHubIssueURL)
		assert.NotNil(t, workflow.GitHubPRURL)
		assert.Equal(t, "https://github.com/owner/repo/pull/456", *workflow.GitHubPRURL)
		assert.NotNil(t, workflow.BranchName)
		assert.Equal(t, "feature/test-branch", *workflow.BranchName)
		assert.NotNil(t, workflow.CleanupStatus)
		assert.Equal(t, "pending", *workflow.CleanupStatus)
	})

	t.Run("Update WorkflowHistories Git fields", func(t *testing.T) {
		newPRURL := "https://github.com/owner/repo/pull/789"
		newCleanupStatus := "completed"

		result := db.Model(&WorkflowHistories{}).
			Where("request_id = ?", "test-request-123").
			Updates(map[string]interface{}{
				"github_pr_url":   newPRURL,
				"cleanup_status":  newCleanupStatus,
				"status":          "completed",
			})
		assert.NoError(t, result.Error)

		// Verify update
		var workflow WorkflowHistories
		db.Where("request_id = ?", "test-request-123").First(&workflow)
		assert.Equal(t, "completed", workflow.Status)
		assert.NotNil(t, workflow.GitHubPRURL)
		assert.Equal(t, newPRURL, *workflow.GitHubPRURL)
		assert.NotNil(t, workflow.CleanupStatus)
		assert.Equal(t, newCleanupStatus, *workflow.CleanupStatus)
	})

	t.Run("Create WorkflowHistories without Git fields", func(t *testing.T) {
		workflow := WorkflowHistories{
			RequestID:      "test-request-456",
			Status:         "pending",
			Tasks:          "Another test task",
			RepositoryName: "another-repo",
			Provider:       "claude",
			CreatedAt:      time.Now(),
		}

		result := db.Create(&workflow)
		assert.NoError(t, result.Error)
		assert.NotZero(t, workflow.ID)

		// Verify Git fields are nil when not set
		var retrieved WorkflowHistories
		db.Where("request_id = ?", "test-request-456").First(&retrieved)
		assert.Nil(t, retrieved.GitHubIssueURL)
		assert.Nil(t, retrieved.GitHubPRURL)
		assert.Nil(t, retrieved.BranchName)
		assert.Nil(t, retrieved.CleanupStatus)
	})

	t.Run("Test table name", func(t *testing.T) {
		workflow := WorkflowHistories{}
		assert.Equal(t, "workflow_histories", workflow.TableName())
	})

	t.Run("Test JSON serialization with Git fields", func(t *testing.T) {
		githubIssueURL := "https://github.com/owner/repo/issues/999"
		branchName := "feature/json-test"

		workflow := WorkflowHistories{
			RequestID:        "json-test-123",
			Status:           "completed",
			Tasks:            "JSON test task",
			RepositoryName:   "json-repo",
			Provider:         "claude",
			CreatedAt:        time.Now(),
			GitHubIssueURL:   &githubIssueURL,
			BranchName:       &branchName,
		}

		// Create and retrieve to test JSON serialization
		db.Create(&workflow)
		var retrieved WorkflowHistories
		db.Where("request_id = ?", "json-test-123").First(&retrieved)

		assert.Equal(t, githubIssueURL, *retrieved.GitHubIssueURL)
		assert.Equal(t, branchName, *retrieved.BranchName)
		assert.Nil(t, retrieved.GitHubPRURL)
		assert.Nil(t, retrieved.CleanupStatus)
	})
}

func TestWorkflowHistoriesBackwardCompatibility(t *testing.T) {
	// Test that existing code still works with new fields
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the WorkflowHistories model
	err = db.AutoMigrate(&WorkflowHistories{})
	assert.NoError(t, err)

	t.Run("Existing workflow creation still works", func(t *testing.T) {
		// Simulate creating a workflow record as existing code would
		workflow := WorkflowHistories{
			RequestID:      "legacy-request-123",
			Status:         "pending",
			Tasks:          "Legacy task",
			RepositoryName: "legacy-repo",
			Provider:       "claude",
			CreatedAt:      time.Now(),
		}

		result := db.Create(&workflow)
		assert.NoError(t, result.Error)
		assert.NotZero(t, workflow.ID)

		// Verify the record exists and has expected values
		var retrieved WorkflowHistories
		db.Where("request_id = ?", "legacy-request-123").First(&retrieved)
		assert.Equal(t, "legacy-request-123", retrieved.RequestID)
		assert.Equal(t, "pending", retrieved.Status)
		assert.Equal(t, "Legacy task", retrieved.Tasks)

		// New fields should be nil/empty
		assert.Nil(t, retrieved.GitHubIssueURL)
		assert.Nil(t, retrieved.GitHubPRURL)
		assert.Nil(t, retrieved.BranchName)
		assert.Nil(t, retrieved.CleanupStatus)
	})

	t.Run("Query without Git fields still works", func(t *testing.T) {
		var workflows []WorkflowHistories
		result := db.Where("status = ?", "pending").Find(&workflows)
		assert.NoError(t, result.Error)
		assert.Len(t, workflows, 1)
		assert.Equal(t, "legacy-request-123", workflows[0].RequestID)
	})

	t.Run("Update without Git fields still works", func(t *testing.T) {
		completedAt := time.Now()
		processingTime := int64(5000)
		result := db.Model(&WorkflowHistories{}).
			Where("request_id = ?", "legacy-request-123").
			Updates(map[string]interface{}{
				"status":            "completed",
				"completed_at":      completedAt,
				"processing_time_ms": processingTime,
			})
		assert.NoError(t, result.Error)

		// Verify update
		var workflow WorkflowHistories
		db.Where("request_id = ?", "legacy-request-123").First(&workflow)
		assert.Equal(t, "completed", workflow.Status)
		assert.NotNil(t, workflow.CompletedAt)
		assert.NotNil(t, workflow.ProcessingTimeMs)
		assert.Equal(t, processingTime, *workflow.ProcessingTimeMs)
	})
}