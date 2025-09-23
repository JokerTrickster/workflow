package usecase

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/request"
	"main/features/task/model/response"
)

type ListTasksUseCase struct {
	listTasksRepo _interface.IListTasksRepository
}

func NewListTasksUseCase(listTasksRepo _interface.IListTasksRepository) _interface.IListTasksUseCase {
	return &ListTasksUseCase{
		listTasksRepo: listTasksRepo,
	}
}

func (uc *ListTasksUseCase) ListTasks(req *request.ListTasksRequest) (*response.ListTasksResponse, error) {
	tasks, totalCount, err := uc.listTasksRepo.GetTasks(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Convert to response format
	var taskResponses []response.TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, response.TaskResponse{
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
		})
	}

	hasMore := int64(req.Page*req.Limit) < totalCount

	return &response.ListTasksResponse{
		Tasks:      taskResponses,
		TotalCount: totalCount,
		Page:       req.Page,
		Limit:      req.Limit,
		HasMore:    hasMore,
	}, nil
}