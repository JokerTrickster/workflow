package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/internal/infrastructure/database"
	mysqlRepo "ai-git-workbench/internal/infrastructure/repositories"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*database.DB, repositories.TaskRepository) {
	// Use test database configuration
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "",
		Name:     "workflow_test",
		Charset:  "utf8mb4",
	}

	db, err := database.NewMySQLConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: MySQL not available: %v", err)
	}

	// Clean up any existing test data
	_, err = db.Exec("TRUNCATE TABLE tasks")
	if err != nil {
		t.Logf("Warning: Could not truncate tasks table: %v", err)
	}
	_, err = db.Exec("TRUNCATE TABLE task_metadata")
	if err != nil {
		t.Logf("Warning: Could not truncate task_metadata table: %v", err)
	}

	repo := mysqlRepo.NewMySQLTaskRepository(db)
	return db, repo
}

func TestTaskRepository_Create(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	task := &entities.Task{
		ID:         uuid.New().String(),
		BranchName: "feature/test",
		Title:      "Test Task",
		Content:    "This is a test task",
		Repository: "test-repo",
		UserID:     "user123",
		Status:     entities.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata: map[string]string{
			"epic":     "test-epic",
			"priority": "high",
		},
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify task was created
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve created task: %v", err)
	}

	// Verify all fields
	if retrievedTask.ID != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, retrievedTask.ID)
	}
	if retrievedTask.Title != task.Title {
		t.Errorf("Expected title %s, got %s", task.Title, retrievedTask.Title)
	}
	if retrievedTask.Status != task.Status {
		t.Errorf("Expected status %s, got %s", task.Status, retrievedTask.Status)
	}
	if retrievedTask.UserID != task.UserID {
		t.Errorf("Expected UserID %s, got %s", task.UserID, retrievedTask.UserID)
	}

	// Verify metadata
	if len(retrievedTask.Metadata) != 2 {
		t.Errorf("Expected 2 metadata items, got %d", len(retrievedTask.Metadata))
	}
	if retrievedTask.Metadata["epic"] != "test-epic" {
		t.Errorf("Expected epic metadata 'test-epic', got %s", retrievedTask.Metadata["epic"])
	}
}

func TestTaskRepository_Update(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create initial task
	task := &entities.Task{
		ID:         uuid.New().String(),
		BranchName: "feature/test",
		Title:      "Original Title",
		Content:    "Original content",
		Repository: "test-repo",
		UserID:     "user123",
		Status:     entities.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Update task
	task.Title = "Updated Title"
	task.Status = entities.TaskStatusProcessing
	task.Metadata = map[string]string{
		"updated": "true",
	}

	err = repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify update
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated task: %v", err)
	}

	if retrievedTask.Title != "Updated Title" {
		t.Errorf("Expected updated title 'Updated Title', got %s", retrievedTask.Title)
	}
	if retrievedTask.Status != entities.TaskStatusProcessing {
		t.Errorf("Expected status %s, got %s", entities.TaskStatusProcessing, retrievedTask.Status)
	}
	if retrievedTask.Metadata["updated"] != "true" {
		t.Errorf("Expected metadata 'updated'='true', got %s", retrievedTask.Metadata["updated"])
	}
}

func TestTaskRepository_List_WithFilters(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test tasks
	tasks := []*entities.Task{
		{
			ID:         uuid.New().String(),
			BranchName: "feature/task1",
			Title:      "Task 1",
			Repository: "repo1",
			UserID:     "user1",
			Status:     entities.TaskStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New().String(),
			BranchName: "feature/task2",
			Title:      "Task 2",
			Repository: "repo1",
			UserID:     "user2",
			Status:     entities.TaskStatusCompleted,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New().String(),
			BranchName: "feature/task3",
			Title:      "Task 3",
			Repository: "repo2",
			UserID:     "user1",
			Status:     entities.TaskStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	// Create all tasks
	for _, task := range tasks {
		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task %s: %v", task.ID, err)
		}
	}

	// Test filter by user
	filter := repositories.TaskFilter{UserID: "user1"}
	results, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list tasks by user: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 tasks for user1, got %d", len(results))
	}

	// Test filter by status
	filter = repositories.TaskFilter{Status: entities.TaskStatusPending}
	results, err = repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list tasks by status: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", len(results))
	}

	// Test filter by repository
	filter = repositories.TaskFilter{Repository: "repo1"}
	results, err = repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list tasks by repository: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 tasks for repo1, got %d", len(results))
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create task
	task := &entities.Task{
		ID:         uuid.New().String(),
		BranchName: "feature/test",
		Title:      "Test Task",
		Repository: "test-repo",
		UserID:     "user123",
		Status:     entities.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Delete task
	err = repo.Delete(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task is deleted
	_, err = repo.GetByID(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task, but got nil")
	}
}

func TestTaskRepository_Count(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test tasks with different users
	for i := 0; i < 5; i++ {
		task := &entities.Task{
			ID:         uuid.New().String(),
			BranchName: "feature/test",
			Title:      "Test Task",
			Repository: "test-repo",
			UserID:     "user1",
			Status:     entities.TaskStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if i < 2 {
			task.UserID = "user2"
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task %d: %v", i, err)
		}
	}

	// Count all tasks
	filter := repositories.TaskFilter{}
	count, err := repo.Count(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to count all tasks: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 total tasks, got %d", count)
	}

	// Count tasks for user1
	filter = repositories.TaskFilter{UserID: "user1"}
	count, err = repo.Count(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to count tasks for user1: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 tasks for user1, got %d", count)
	}

	// Count tasks for user2
	filter = repositories.TaskFilter{UserID: "user2"}
	count, err = repo.Count(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to count tasks for user2: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 tasks for user2, got %d", count)
	}
}