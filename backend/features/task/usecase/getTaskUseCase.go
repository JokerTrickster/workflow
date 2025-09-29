package usecase

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/response"
)

type GetTaskUseCase struct {
	getTaskRepo _interface.IGetTaskStatusRepository
}

func NewGetTaskUseCase(getTaskRepo _interface.IGetTaskStatusRepository) _interface.IGetTaskUseCase {
	return &GetTaskUseCase{
		getTaskRepo: getTaskRepo,
	}
}

func (uc *GetTaskUseCase) GetTask(requestID string) (*response.TaskResponse, error) {
	task, err := uc.getTaskRepo.GetTaskByRequestID(requestID)
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