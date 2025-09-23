package usecase

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/response"
)

type GetTaskStatusUseCase struct {
	getTaskStatusRepo _interface.IGetTaskStatusRepository
}

func NewGetTaskStatusUseCase(getTaskStatusRepo _interface.IGetTaskStatusRepository) _interface.IGetTaskStatusUseCase {
	return &GetTaskStatusUseCase{
		getTaskStatusRepo: getTaskStatusRepo,
	}
}

func (uc *GetTaskStatusUseCase) GetTaskStatus(requestID string) (*response.TaskResponse, error) {
	task, err := uc.getTaskStatusRepo.GetTaskByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &response.TaskResponse{
		ID:               task.ID,
		RequestID:        task.RequestID,
		Status:           task.Status,
		Tasks:            task.Tasks,
		RepositoryName:   task.RepositoryName,
		WorkingDir:       task.WorkingDir,
		Cmd:              task.Cmd,
		Provider:         task.Provider,
		Interactive:      task.Interactive,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
		CompletedAt:      task.CompletedAt,
		ProcessingTimeMs: task.ProcessingTimeMs,
		Result:           task.Result,
		Error:            task.Error,
	}, nil
}