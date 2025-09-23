package usecase

import (
	"fmt"
	"time"
	_interface "main/features/task/model/interface"
	"main/features/task/model/response"
	"main/utils/db"
)

type CancelTaskUseCase struct {
	cancelTaskRepo _interface.ICancelTaskRepository
}

func NewCancelTaskUseCase(cancelTaskRepo _interface.ICancelTaskRepository) _interface.ICancelTaskUseCase {
	return &CancelTaskUseCase{
		cancelTaskRepo: cancelTaskRepo,
	}
}

func (uc *CancelTaskUseCase) CancelTask(requestID string) (*response.CancelTaskResponse, error) {
	// Check if task exists
	task, err := uc.cancelTaskRepo.GetTaskByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task can be cancelled
	if task.Status == db.TaskStatusCompleted || task.Status == db.TaskStatusCancelled {
		return nil, fmt.Errorf("task cannot be cancelled, current status: %s", task.Status)
	}

	// Update status to cancelled
	if err := uc.cancelTaskRepo.UpdateTaskStatus(requestID, db.TaskStatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return &response.CancelTaskResponse{
		RequestID:   requestID,
		Status:      db.TaskStatusCancelled,
		Message:     "Task cancelled successfully",
		CancelledAt: time.Now().Format(time.RFC3339),
	}, nil
}