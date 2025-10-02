package usecase

import (
	"main/features/epic-tasks/model/request"
	"main/features/epic-tasks/model/response"
	"main/features/epic-tasks/repository"
)

type EpicTaskUseCase struct {
	repo *repository.EpicTaskRepository
}

func NewEpicTaskUseCase() *EpicTaskUseCase {
	return &EpicTaskUseCase{
		repo: repository.NewEpicTaskRepository(),
	}
}

func (uc *EpicTaskUseCase) GetAllTasks(repository string) (*response.EpicTasksListResponse, error) {
	// Use default repository if not provided
	if repository == "" {
		repository = "workflow"
	}

	// Validate repository
	if err := uc.repo.ValidateRepository(repository); err != nil {
		return nil, err
	}

	tasks, err := uc.repo.GetAllTasks(repository)
	if err != nil {
		return nil, err
	}

	return &response.EpicTasksListResponse{Tasks: tasks}, nil
}

func (uc *EpicTaskUseCase) GetTask(repository, taskID string) (*response.EpicTaskResponse, error) {
	// Use default repository if not provided
	if repository == "" {
		repository = "workflow"
	}

	// Validate repository
	if err := uc.repo.ValidateRepository(repository); err != nil {
		return nil, err
	}

	task, err := uc.repo.GetTask(repository, taskID)
	if err != nil {
		return nil, err
	}

	return &response.EpicTaskResponse{Task: *task}, nil
}

func (uc *EpicTaskUseCase) CreateTask(req *request.CreateEpicTaskRequest) (*response.EpicTaskResponse, error) {
	// Use default repository if not provided
	if req.Repository == "" {
		req.Repository = "workflow"
	}

	// Validate repository
	if err := uc.repo.ValidateRepository(req.Repository); err != nil {
		return nil, err
	}

	if err := uc.repo.CreateTask(req.Repository, &req.Task); err != nil {
		return nil, err
	}

	return &response.EpicTaskResponse{Task: req.Task}, nil
}

func (uc *EpicTaskUseCase) UpdateTask(req *request.UpdateEpicTaskRequest) (*response.EpicTaskResponse, error) {
	// Use default repository if not provided
	if req.Repository == "" {
		req.Repository = "workflow"
	}

	// Validate repository
	if err := uc.repo.ValidateRepository(req.Repository); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateTask(req.Repository, &req.Task); err != nil {
		return nil, err
	}

	return &response.EpicTaskResponse{Task: req.Task}, nil
}

func (uc *EpicTaskUseCase) DeleteTask(repository, taskID string) (*response.DeleteEpicTaskResponse, error) {
	// Use default repository if not provided
	if repository == "" {
		repository = "workflow"
	}

	// Validate repository
	if err := uc.repo.ValidateRepository(repository); err != nil {
		return nil, err
	}

	if err := uc.repo.DeleteTask(repository, taskID); err != nil {
		return nil, err
	}

	return &response.DeleteEpicTaskResponse{Success: true}, nil
}
