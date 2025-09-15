package tests

import (
	"testing"
	"time"

	"ai-git-workbench/internal/domain/entities"
)

func TestTaskEntity_Creation(t *testing.T) {
	task := &entities.Task{
		ID:         "test-123",
		BranchName: "feature/test",
		Title:      "Test Task",
		Content:    "This is a test task",
		Repository: "test-repo",
		UserID:     "user123",
		Status:     entities.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata: map[string]string{
			"epic": "test-epic",
		},
	}

	if task.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", task.ID)
	}
	if task.Status != entities.TaskStatusPending {
		t.Errorf("Expected status 'pending', got %s", task.Status)
	}
}

func TestTaskEntity_StatusTransitions(t *testing.T) {
	task := &entities.Task{
		Status: entities.TaskStatusPending,
	}

	// Valid transition: pending -> processing
	if !task.CanTransitionTo(entities.TaskStatusProcessing) {
		t.Error("Should be able to transition from pending to processing")
	}

	// Valid transition: pending -> cancelled
	if !task.CanTransitionTo(entities.TaskStatusCancelled) {
		t.Error("Should be able to transition from pending to cancelled")
	}

	// Invalid transition: pending -> completed
	if task.CanTransitionTo(entities.TaskStatusCompleted) {
		t.Error("Should not be able to transition from pending to completed")
	}

	// Test UpdateStatus
	if !task.UpdateStatus(entities.TaskStatusProcessing) {
		t.Error("UpdateStatus should succeed for valid transition")
	}
	if task.Status != entities.TaskStatusProcessing {
		t.Error("Status should be updated to processing")
	}

	// Test invalid transition
	if task.UpdateStatus(entities.TaskStatusPending) {
		t.Error("UpdateStatus should fail for invalid transition")
	}
}

func TestIsValidStatus(t *testing.T) {
	validStatuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
	invalidStatuses := []string{"invalid", "unknown", "", "PENDING"}

	for _, status := range validStatuses {
		if !entities.IsValidStatus(status) {
			t.Errorf("Status '%s' should be valid", status)
		}
	}

	for _, status := range invalidStatuses {
		if entities.IsValidStatus(status) {
			t.Errorf("Status '%s' should be invalid", status)
		}
	}
}