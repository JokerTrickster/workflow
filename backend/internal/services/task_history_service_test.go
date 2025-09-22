package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database/models"
)

// setupTaskHistoryTestDB creates an in-memory SQLite database for testing
func setupTaskHistoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the workflow_histories table
	err = db.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	return db
}

// seedTestData inserts test data into the database
func seedTestData(t *testing.T, db *gorm.DB) {
	testData := []models.WorkflowHistory{
		{
			RequestID:      "req-1",
			Status:         "completed",
			Tasks:          "Test task 1",
			RepositoryName: "test-repo",
			CreatedAt:      time.Now().Add(-3 * time.Hour),
		},
		{
			RequestID:      "req-2",
			Status:         "pending",
			Tasks:          "Test task 2",
			RepositoryName: "test-repo",
			CreatedAt:      time.Now().Add(-2 * time.Hour),
		},
		{
			RequestID:      "req-3",
			Status:         "failed",
			Tasks:          "Test task 3",
			RepositoryName: "test-repo",
			CreatedAt:      time.Now().Add(-1 * time.Hour),
		},
		{
			RequestID:      "req-4",
			Status:         "completed",
			Tasks:          "Test task 4",
			RepositoryName: "other-repo",
			CreatedAt:      time.Now().Add(-30 * time.Minute),
		},
	}

	for _, data := range testData {
		err := db.Create(&data).Error
		require.NoError(t, err)
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid parameters",
			page:          1,
			limit:         20,
			expectedError: false,
		},
		{
			name:          "valid max limit",
			page:          1,
			limit:         100,
			expectedError: false,
		},
		{
			name:          "invalid page zero",
			page:          0,
			limit:         20,
			expectedError: true,
			errorContains: "page must be greater than 0",
		},
		{
			name:          "invalid negative page",
			page:          -1,
			limit:         20,
			expectedError: true,
			errorContains: "page must be greater than 0",
		},
		{
			name:          "invalid limit zero",
			page:          1,
			limit:         0,
			expectedError: true,
			errorContains: "limit must be greater than 0",
		},
		{
			name:          "invalid limit too high",
			page:          1,
			limit:         101,
			expectedError: true,
			errorContains: "limit exceeds maximum allowed value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePaginationParams(tt.page, tt.limit)

			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.page, result.Page)
				assert.Equal(t, tt.limit, result.Limit)
			}
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "valid repository name",
			repositoryName: "test-repo",
			expectedError:  false,
		},
		{
			name:           "empty repository name",
			repositoryName: "",
			expectedError:  true,
			errorContains:  "repository name is required",
		},
		{
			name:           "repository name too long",
			repositoryName: string(make([]byte, 256)), // 256 characters
			expectedError:  true,
			errorContains:  "repository name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepositoryName(tt.repositoryName)

			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tt.errorContains)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestTaskHistoryService_GetTaskHistory(t *testing.T) {
	db := setupTaskHistoryTestDB(t)
	seedTestData(t, db)
	service := NewTaskHistoryService(db)
	ctx := context.Background()

	t.Run("successful retrieval with pagination", func(t *testing.T) {
		params := &PaginationParams{Page: 1, Limit: 20}
		result, err := service.GetTaskHistory(ctx, "test-repo", params)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 3) // 3 tasks for test-repo
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 3, result.Pagination.Total)
		assert.Equal(t, 1, result.Pagination.TotalPages)

		// Verify tasks are sorted by created_at DESC (newest first)
		assert.True(t, result.Data[0].CreatedAt.After(result.Data[1].CreatedAt))
		assert.True(t, result.Data[1].CreatedAt.After(result.Data[2].CreatedAt))
	})

	t.Run("pagination with limit", func(t *testing.T) {
		params := &PaginationParams{Page: 1, Limit: 2}
		result, err := service.GetTaskHistory(ctx, "test-repo", params)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2) // Limited to 2 tasks
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 2, result.Pagination.Limit)
		assert.Equal(t, 3, result.Pagination.Total)
		assert.Equal(t, 2, result.Pagination.TotalPages)
	})

	t.Run("second page", func(t *testing.T) {
		params := &PaginationParams{Page: 2, Limit: 2}
		result, err := service.GetTaskHistory(ctx, "test-repo", params)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1) // 1 task on second page
		assert.Equal(t, 2, result.Pagination.Page)
		assert.Equal(t, 2, result.Pagination.Limit)
		assert.Equal(t, 3, result.Pagination.Total)
		assert.Equal(t, 2, result.Pagination.TotalPages)
	})

	t.Run("empty repository", func(t *testing.T) {
		params := &PaginationParams{Page: 1, Limit: 20}
		result, err := service.GetTaskHistory(ctx, "nonexistent-repo", params)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 0)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Total)
		assert.Equal(t, 0, result.Pagination.TotalPages)
	})

	t.Run("invalid repository name", func(t *testing.T) {
		params := &PaginationParams{Page: 1, Limit: 20}
		result, err := service.GetTaskHistory(ctx, "", params)

		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Message, "repository name is required")
	})

	t.Run("page exceeds available pages", func(t *testing.T) {
		params := &PaginationParams{Page: 10, Limit: 20}
		result, err := service.GetTaskHistory(ctx, "test-repo", params)

		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Message, "page exceeds available pages")
	})

	t.Run("repository filtering", func(t *testing.T) {
		params := &PaginationParams{Page: 1, Limit: 20}
		result, err := service.GetTaskHistory(ctx, "other-repo", params)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1) // 1 task for other-repo
		assert.Equal(t, "other-repo", result.Data[0].RepositoryName)
	})
}

func TestTaskHistoryService_CheckRepositoryExists(t *testing.T) {
	db := setupTaskHistoryTestDB(t)
	seedTestData(t, db)
	service := NewTaskHistoryService(db)
	ctx := context.Background()

	t.Run("existing repository", func(t *testing.T) {
		exists, err := service.CheckRepositoryExists(ctx, "test-repo")

		assert.Nil(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing repository", func(t *testing.T) {
		exists, err := service.CheckRepositoryExists(ctx, "nonexistent-repo")

		assert.Nil(t, err)
		assert.False(t, exists)
	})
}