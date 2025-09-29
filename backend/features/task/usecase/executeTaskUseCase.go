package usecase

import (
	"fmt"
	"time"
	_interface "main/features/task/model/interface"
	"main/features/task/model/response"
	"main/utils"
	"main/utils/db"
)

type ExecuteTaskUseCase struct {
	executeTaskRepo _interface.IExecuteTaskRepository
}

func NewExecuteTaskUseCase(executeTaskRepo _interface.IExecuteTaskRepository) _interface.IExecuteTaskUseCase {
	return &ExecuteTaskUseCase{
		executeTaskRepo: executeTaskRepo,
	}
}

func (uc *ExecuteTaskUseCase) ExecuteTask(requestID string) (*response.ExecuteTaskResponse, error) {
	// Check if task exists
	task, err := uc.executeTaskRepo.GetTaskByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task is in pending status
	if task.Status != db.TaskStatusPending {
		return nil, fmt.Errorf("task is not in pending status, current status: %s", task.Status)
	}

	// Update status to processing
	if err := uc.executeTaskRepo.UpdateTaskStatus(requestID, db.TaskStatusProcessing); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Create message for RabbitMQ queue
	workingDir := ""
	if task.WorkingDir != nil {
		workingDir = *task.WorkingDir
	}

	cmd := ""
	if task.Cmd != nil {
		cmd = *task.Cmd
	}

	taskMessage := &utils.TaskMessage{
		RequestID:      requestID,
		Tasks:          task.Tasks,
		RepositoryName: task.RepositoryName,
		WorkingDir:     workingDir,
		Interactive:    task.Interactive,
		Cmd:            cmd,
		ContinueTask:   false, // Default to false, can be made configurable
		Provider:       task.Provider,
	}

	// Send task to RabbitMQ queue for processing
	if err := utils.PublishTask(taskMessage); err != nil {
		// Log the error but don't fail the request - the task status is already updated
		fmt.Printf("Failed to publish task to RabbitMQ: %v\n", err)
		// Note: In production, you might want to implement retry logic or dead letter queues
	}

	return &response.ExecuteTaskResponse{
		RequestID: requestID,
		Status:    db.TaskStatusProcessing,
		Message:   "Task execution started successfully",
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}