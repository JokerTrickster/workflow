package dto_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
)

func TestCreateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request dto.CreateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
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
			name: "Empty user ID should fail",
			request: dto.CreateTaskRequest{
				UserID:      "",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "user_id",
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
			name: "Title too long should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       string(make([]byte, 501)), // 501 characters
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "Description too long should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: string(make([]byte, 5001)), // 5001 characters
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "description",
		},
		{
			name: "Empty repository should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "",
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
		{
			name: "Epic too long should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        string(make([]byte, 256)), // 256 characters
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "epic",
		},
		{
			name: "Empty branch should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "This is a test task",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "",
			},
			wantErr: true,
			errMsg:  "branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestCreateTaskRequest_ToValueObjects(t *testing.T) {
	tests := []struct {
		name    string
		request dto.CreateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid conversion",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Repository:  "owner/repo",
				Branch:      "feature/test",
			},
			wantErr: false,
		},
		{
			name: "Invalid user ID should fail",
			request: dto.CreateTaskRequest{
				UserID:      "", // Invalid
				Repository:  "owner/repo",
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "user ID",
		},
		{
			name: "Invalid repository should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Repository:  "invalid-repo", // Invalid format
				Branch:      "feature/test",
			},
			wantErr: true,
			errMsg:  "repository",
		},
		{
			name: "Invalid branch should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Repository:  "owner/repo",
				Branch:      "", // Invalid
			},
			wantErr: true,
			errMsg:  "branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, repository, branch, err := tt.request.ToValueObjects()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ToValueObjects() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ToValueObjects() error = %v, expected to contain %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ToValueObjects() unexpected error = %v", err)
				return
			}

			if userID.Value() != tt.request.UserID {
				t.Errorf("ToValueObjects() userID = %v, expected %v", userID.Value(), tt.request.UserID)
			}

			if repository.Value() != tt.request.Repository {
				t.Errorf("ToValueObjects() repository = %v, expected %v", repository.Value(), tt.request.Repository)
			}

			if branch.Value() != tt.request.Branch {
				t.Errorf("ToValueObjects() branch = %v, expected %v", branch.Value(), tt.request.Branch)
			}
		})
	}
}

func TestToGetTaskResponse(t *testing.T) {
	// Create a test task entity
	userID, _ := valueobjects.NewUserID("user123")
	repository, _ := valueobjects.NewRepositoryPath("owner/repo")
	branch, _ := valueobjects.NewBranchName("feature/test")

	task, err := entities.CreateTask(
		userID,
		"Test Task",
		"This is a test task",
		repository,
		"epic-1",
		branch,
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Add some metadata
	task.SetMetadata("key1", "value1")
	task.SetMetadata("key2", "value2")

	// Convert to DTO
	response := dto.ToGetTaskResponse(task)

	// Validate conversion
	if response.TaskID != task.ID().Value() {
		t.Errorf("ToGetTaskResponse() TaskID = %v, expected %v", response.TaskID, task.ID().Value())
	}

	if response.UserID != task.UserID().Value() {
		t.Errorf("ToGetTaskResponse() UserID = %v, expected %v", response.UserID, task.UserID().Value())
	}

	if response.Title != task.Title() {
		t.Errorf("ToGetTaskResponse() Title = %v, expected %v", response.Title, task.Title())
	}

	if response.Description != task.Description() {
		t.Errorf("ToGetTaskResponse() Description = %v, expected %v", response.Description, task.Description())
	}

	if response.Status != task.Status().Value() {
		t.Errorf("ToGetTaskResponse() Status = %v, expected %v", response.Status, task.Status().Value())
	}

	if response.Repository != task.Repository().Value() {
		t.Errorf("ToGetTaskResponse() Repository = %v, expected %v", response.Repository, task.Repository().Value())
	}

	if response.Epic != task.Epic() {
		t.Errorf("ToGetTaskResponse() Epic = %v, expected %v", response.Epic, task.Epic())
	}

	if response.Branch != task.Branch().Value() {
		t.Errorf("ToGetTaskResponse() Branch = %v, expected %v", response.Branch, task.Branch().Value())
	}

	if response.TokensUsed != task.TokensUsed() {
		t.Errorf("ToGetTaskResponse() TokensUsed = %v, expected %v", response.TokensUsed, task.TokensUsed())
	}

	if response.Version != task.Version() {
		t.Errorf("ToGetTaskResponse() Version = %v, expected %v", response.Version, task.Version())
	}

	// Validate timestamps
	if response.CreatedAt == "" {
		t.Errorf("ToGetTaskResponse() CreatedAt is empty")
	}

	if response.UpdatedAt == "" {
		t.Errorf("ToGetTaskResponse() UpdatedAt is empty")
	}

	// Validate metadata
	if len(response.Metadata) != 2 {
		t.Errorf("ToGetTaskResponse() Metadata count = %v, expected 2", len(response.Metadata))
	}

	if response.Metadata["key1"] != "value1" {
		t.Errorf("ToGetTaskResponse() Metadata[key1] = %v, expected value1", response.Metadata["key1"])
	}

	if response.Metadata["key2"] != "value2" {
		t.Errorf("ToGetTaskResponse() Metadata[key2] = %v, expected value2", response.Metadata["key2"])
	}

	// StartedAt and CompletedAt should be nil for new task
	if response.StartedAt != nil {
		t.Errorf("ToGetTaskResponse() StartedAt = %v, expected nil", response.StartedAt)
	}

	if response.CompletedAt != nil {
		t.Errorf("ToGetTaskResponse() CompletedAt = %v, expected nil", response.CompletedAt)
	}
}

func TestToCreateTaskResponse(t *testing.T) {
	// Create a test task entity
	userID, _ := valueobjects.NewUserID("user123")
	repository, _ := valueobjects.NewRepositoryPath("owner/repo")
	branch, _ := valueobjects.NewBranchName("feature/test")

	task, err := entities.CreateTask(
		userID,
		"Test Task",
		"This is a test task",
		repository,
		"epic-1",
		branch,
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Convert to DTO
	response := dto.ToCreateTaskResponse(task)

	// Validate conversion
	if response.TaskID != task.ID().Value() {
		t.Errorf("ToCreateTaskResponse() TaskID = %v, expected %v", response.TaskID, task.ID().Value())
	}

	if response.Status != task.Status().Value() {
		t.Errorf("ToCreateTaskResponse() Status = %v, expected %v", response.Status, task.Status().Value())
	}

	if response.CreatedAt == "" {
		t.Errorf("ToCreateTaskResponse() CreatedAt is empty")
	}

	// Validate timestamp format
	_, err = time.Parse(time.RFC3339, response.CreatedAt)
	if err != nil {
		t.Errorf("ToCreateTaskResponse() CreatedAt format is invalid: %v", err)
	}
}

func TestToListTasksResponse(t *testing.T) {
	// Create test tasks
	userID, _ := valueobjects.NewUserID("user123")
	repository, _ := valueobjects.NewRepositoryPath("owner/repo")
	branch, _ := valueobjects.NewBranchName("feature/test")

	var tasks []*entities.Task
	for i := 0; i < 3; i++ {
		task, err := entities.CreateTask(
			userID,
			fmt.Sprintf("Test Task %d", i+1),
			"This is a test task",
			repository,
			"epic-1",
			branch,
		)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}
		tasks = append(tasks, task)
	}

	// Test conversion
	total := 10
	limit := 5
	offset := 2

	response := dto.ToListTasksResponse(tasks, total, limit, offset)

	// Validate response
	if len(response.Tasks) != len(tasks) {
		t.Errorf("ToListTasksResponse() task count = %v, expected %v", len(response.Tasks), len(tasks))
	}

	if response.Total != total {
		t.Errorf("ToListTasksResponse() Total = %v, expected %v", response.Total, total)
	}

	if response.Limit != limit {
		t.Errorf("ToListTasksResponse() Limit = %v, expected %v", response.Limit, limit)
	}

	if response.Offset != offset {
		t.Errorf("ToListTasksResponse() Offset = %v, expected %v", response.Offset, offset)
	}

	// HasMore should be true when offset + task count < total
	expectedHasMore := offset+len(tasks) < total
	if response.HasMore != expectedHasMore {
		t.Errorf("ToListTasksResponse() HasMore = %v, expected %v", response.HasMore, expectedHasMore)
	}

	// Validate individual task conversions
	for i, taskResponse := range response.Tasks {
		if taskResponse.TaskID != tasks[i].ID().Value() {
			t.Errorf("ToListTasksResponse() Tasks[%d].TaskID = %v, expected %v", i, taskResponse.TaskID, tasks[i].ID().Value())
		}

		if taskResponse.Title != tasks[i].Title() {
			t.Errorf("ToListTasksResponse() Tasks[%d].Title = %v, expected %v", i, taskResponse.Title, tasks[i].Title())
		}
	}
}

func TestValidationError(t *testing.T) {
	err := &dto.ValidationError{
		Field:   "title",
		Message: "is required",
	}

	expected := "title is required"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, expected %v", err.Error(), expected)
	}
}

// Helper function
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}