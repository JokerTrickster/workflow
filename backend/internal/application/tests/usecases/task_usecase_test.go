package usecases_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/application/interfaces"
	"ai-git-workbench/internal/application/services"
	"ai-git-workbench/internal/application/usecases"
	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
	domainServices "ai-git-workbench/internal/domain/services"
	"ai-git-workbench/internal/infrastructure/database"
	mysqlRepo "ai-git-workbench/internal/infrastructure/repositories"
)

func setupTestEnvironment(t *testing.T) (interfaces.TaskUsecase, context.Context) {
	// For integration tests, we would need a real database connection
	// For now, we'll skip database-dependent tests and focus on unit tests
	// TODO: Add integration tests with test database setup
	t.Skip("Integration tests require database setup - implement with test database")
	
	return nil, context.Background()
}

func TestTaskUsecase_CreateTask(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	tests := []struct {
		name    string
		request dto.CreateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid task creation",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: false,
		},
		{
			name: "Empty title should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "Invalid repository format should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "invalid-repo-format",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "repository",
		},
		{
			name: "Empty epic should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.CreateTask(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateTask() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateTask() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateTask() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("CreateTask() response is nil")
				return
			}

			if response.TaskID == "" {
				t.Errorf("CreateTask() TaskID is empty")
			}

			if response.Status != "pending" {
				t.Errorf("CreateTask() Status = %v, expected pending", response.Status)
			}

			if response.CreatedAt == "" {
				t.Errorf("CreateTask() CreatedAt is empty")
			}
		})
	}
}

func TestTaskUsecase_GetTask(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create a test task first
	createReq := dto.CreateTaskRequest{
		UserID:      "user123",
		Title:       "Test Task",
		Description: "This is a test task",
		Repository:  "owner/repo",
		Epic:        "epic-1",
		Branch:      "feature/test",
	}

	createResp, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid task retrieval",
			taskID:  createResp.TaskID,
			userID:  "user123",
			wantErr: false,
		},
		{
			name:    "Invalid task ID should fail",
			taskID:  "invalid-task-id",
			userID:  "user123",
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:    "Unauthorized access should fail",
			taskID:  createResp.TaskID,
			userID:  "different-user",
			wantErr: true,
			errMsg:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.GetTask(ctx, tt.taskID, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTask() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetTask() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("GetTask() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("GetTask() response is nil")
				return
			}

			if response.TaskID != createResp.TaskID {
				t.Errorf("GetTask() TaskID = %v, expected %v", response.TaskID, createResp.TaskID)
			}

			if response.UserID != "user123" {
				t.Errorf("GetTask() UserID = %v, expected user123", response.UserID)
			}

			if response.Title != "Test Task" {
				t.Errorf("GetTask() Title = %v, expected Test Task", response.Title)
			}
		})
	}
}

func TestTaskUsecase_UpdateTask(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create a test task first
	createReq := dto.CreateTaskRequest{
		UserID:      "user123",
		Title:       "Test Task",
		Description: "This is a test task",
		Repository:  "owner/repo",
		Epic:        "epic-1",
		Branch:      "feature/test",
	}

	createResp, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Get the task to get the current version
	getResp, err := usecase.GetTask(ctx, createResp.TaskID, "user123")
	if err != nil {
		t.Fatalf("Failed to get test task: %v", err)
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		request dto.UpdateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "Valid task update",
			taskID: createResp.TaskID,
			userID: "user123",
			request: dto.UpdateTaskRequest{
				Title:   stringPtr("Updated Task Title"),
				Version: getResp.Version,
			},
			wantErr: false,
		},
		{
			name:   "Stale version should fail",
			taskID: createResp.TaskID,
			userID: "user123",
			request: dto.UpdateTaskRequest{
				Title:   stringPtr("Another Update"),
				Version: getResp.Version - 1, // Stale version
			},
			wantErr: true,
			errMsg:  "modified by another process",
		},
		{
			name:   "Unauthorized update should fail",
			taskID: createResp.TaskID,
			userID: "different-user",
			request: dto.UpdateTaskRequest{
				Title:   stringPtr("Unauthorized Update"),
				Version: getResp.Version,
			},
			wantErr: true,
			errMsg:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.UpdateTask(ctx, tt.taskID, tt.userID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateTask() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("UpdateTask() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateTask() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("UpdateTask() response is nil")
				return
			}

			if tt.request.Title != nil && response.Title != *tt.request.Title {
				t.Errorf("UpdateTask() Title = %v, expected %v", response.Title, *tt.request.Title)
			}

			if response.Version <= getResp.Version {
				t.Errorf("UpdateTask() Version = %v, expected > %v", response.Version, getResp.Version)
			}
		})
	}
}

func TestTaskUsecase_ListTasks(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create multiple test tasks
	tasks := []dto.CreateTaskRequest{
		{
			UserID:      "user123",
			Title:       "Task 1",
			Description: "First task",
			Repository:  "owner/repo1",
			Epic:        "epic-1",
			Branch:      "feature/test1",
		},
		{
			UserID:      "user123",
			Title:       "Task 2",
			Description: "Second task",
			Repository:  "owner/repo2",
			Epic:        "epic-2",
			Branch:      "feature/test2",
		},
		{
			UserID:      "user456",
			Title:       "Task 3",
			Description: "Third task",
			Repository:  "owner/repo3",
			Epic:        "epic-3",
			Branch:      "feature/test3",
		},
	}

	for _, task := range tasks {
		_, err := usecase.CreateTask(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}
	}

	tests := []struct {
		name         string
		request      dto.ListTasksRequest
		expectedCount int
		wantErr      bool
	}{
		{
			name: "List all tasks for user123",
			request: dto.ListTasksRequest{
				UserID: stringPtr("user123"),
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "List tasks with limit",
			request: dto.ListTasksRequest{
				UserID: stringPtr("user123"),
				Limit:  1,
				Offset: 0,
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name: "List tasks with filter by repository",
			request: dto.ListTasksRequest{
				UserID:     stringPtr("user123"),
				Repository: stringPtr("owner/repo1"),
				Limit:      10,
				Offset:     0,
			},
			expectedCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.ListTasks(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListTasks() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ListTasks() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("ListTasks() response is nil")
				return
			}

			if len(response.Tasks) != tt.expectedCount {
				t.Errorf("ListTasks() task count = %v, expected %v", len(response.Tasks), tt.expectedCount)
			}

			if response.Limit != tt.request.Limit {
				t.Errorf("ListTasks() Limit = %v, expected %v", response.Limit, tt.request.Limit)
			}

			if response.Offset != tt.request.Offset {
				t.Errorf("ListTasks() Offset = %v, expected %v", response.Offset, tt.request.Offset)
			}
		})
	}
}

func TestTaskUsecase_CancelTask(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create a test task first
	createReq := dto.CreateTaskRequest{
		UserID:      "user123",
		Title:       "Test Task",
		Description: "This is a test task",
		Repository:  "owner/repo",
		Epic:        "epic-1",
		Branch:      "feature/test",
	}

	createResp, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		request dto.TaskActionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "Valid task cancellation",
			taskID: createResp.TaskID,
			userID: "user123",
			request: dto.TaskActionRequest{
				Reason: "Test cancellation",
			},
			wantErr: false,
		},
		{
			name:   "Unauthorized cancellation should fail",
			taskID: createResp.TaskID,
			userID: "different-user",
			request: dto.TaskActionRequest{
				Reason: "Unauthorized cancellation",
			},
			wantErr: true,
			errMsg:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.CancelTask(ctx, tt.taskID, tt.userID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CancelTask() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("CancelTask() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CancelTask() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("CancelTask() response is nil")
				return
			}

			if response.Status != "cancelled" {
				t.Errorf("CancelTask() Status = %v, expected cancelled", response.Status)
			}

			if response.TaskID != tt.taskID {
				t.Errorf("CancelTask() TaskID = %v, expected %v", response.TaskID, tt.taskID)
			}
		})
	}
}

func TestTaskUsecase_ResumeTask(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create and cancel a test task first
	createReq := dto.CreateTaskRequest{
		UserID:      "user123",
		Title:       "Test Task",
		Description: "This is a test task",
		Repository:  "owner/repo",
		Epic:        "epic-1",
		Branch:      "feature/test",
	}

	createResp, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Cancel the task first
	_, err = usecase.CancelTask(ctx, createResp.TaskID, "user123", dto.TaskActionRequest{
		Reason: "For testing resume",
	})
	if err != nil {
		t.Fatalf("Failed to cancel test task: %v", err)
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		request dto.TaskActionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "Valid task resume",
			taskID: createResp.TaskID,
			userID: "user123",
			request: dto.TaskActionRequest{
				Reason: "Test resume",
			},
			wantErr: false,
		},
		{
			name:   "Unauthorized resume should fail",
			taskID: createResp.TaskID,
			userID: "different-user",
			request: dto.TaskActionRequest{
				Reason: "Unauthorized resume",
			},
			wantErr: true,
			errMsg:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := usecase.ResumeTask(ctx, tt.taskID, tt.userID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResumeTask() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ResumeTask() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ResumeTask() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Errorf("ResumeTask() response is nil")
				return
			}

			if response.Status != "pending" {
				t.Errorf("ResumeTask() Status = %v, expected pending", response.Status)
			}

			if response.TaskID != tt.taskID {
				t.Errorf("ResumeTask() TaskID = %v, expected %v", response.TaskID, tt.taskID)
			}
		})
	}
}

func TestTaskUsecase_GetTaskStatistics(t *testing.T) {
	usecase, ctx := setupTestEnvironment(t)

	// Create some test tasks with different statuses
	userID := "user123"
	
	// Create tasks
	for i := 0; i < 3; i++ {
		createReq := dto.CreateTaskRequest{
			UserID:      userID,
			Title:       fmt.Sprintf("Task %d", i+1),
			Description: "Test task",
			Repository:  "owner/repo",
			Epic:        "epic-1",
			Branch:      "feature/test",
		}
		_, err := usecase.CreateTask(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}
	}

	// Wait a moment to ensure tasks are created
	time.Sleep(10 * time.Millisecond)

	response, err := usecase.GetTaskStatistics(ctx, userID)
	if err != nil {
		t.Errorf("GetTaskStatistics() unexpected error = %v", err)
		return
	}

	if response == nil {
		t.Errorf("GetTaskStatistics() response is nil")
		return
	}

	if response.UserID != userID {
		t.Errorf("GetTaskStatistics() UserID = %v, expected %v", response.UserID, userID)
	}

	if response.TotalTasks < 3 {
		t.Errorf("GetTaskStatistics() TotalTasks = %v, expected >= 3", response.TotalTasks)
	}

	if response.PendingTasks < 3 {
		t.Errorf("GetTaskStatistics() PendingTasks = %v, expected >= 3", response.PendingTasks)
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}