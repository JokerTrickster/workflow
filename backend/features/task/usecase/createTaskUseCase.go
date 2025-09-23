package usecase

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/request"
	"main/features/task/model/response"
	"main/utils/db"
	"time"

	"github.com/google/uuid"
)

type CreateTaskUseCase struct {
	createTaskRepo _interface.ICreateTaskRepository
}

func NewCreateTaskUseCase(createTaskRepo _interface.ICreateTaskRepository) _interface.ICreateTaskUseCase {
	return &CreateTaskUseCase{
		createTaskRepo: createTaskRepo,
	}
}

func (uc *CreateTaskUseCase) CreateTask(req *request.CreateTaskRequest) (*response.CreateTaskResponse, error) {
	// Generate unique request ID
	requestID := fmt.Sprintf("task-%s", uuid.New().String())

	// Create task record
	now := time.Now()
	task := &db.Task{
		RequestID:      requestID,
		Status:         db.TaskStatusPending,
		Tasks:          req.Tasks,
		RepositoryName: req.RepositoryName,
		WorkingDir:     &req.WorkingDir,
		Cmd:            &req.Cmd,
		Provider:       req.Provider,
		Interactive:    req.Interactive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.createTaskRepo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &response.CreateTaskResponse{
		RequestID: requestID,
		Status:    db.TaskStatusPending,
		Message:   "Task created successfully",
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}