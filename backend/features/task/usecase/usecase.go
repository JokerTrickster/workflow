package usecase

import (
	"fmt"
	_interface "main/features/task/model/interface"
	"main/features/task/model/request"
	"main/features/task/model/response"
	"main/utils"
	"main/utils/db"
	"time"

	"github.com/google/uuid"
)

type TaskUseCase struct {
	taskRepo _interface.ITaskRepository
}

func NewTaskUseCase() _interface.ITaskUseCase {
	return &TaskUseCase{
		taskRepo: nil, // Repository는 런타임에 주입됨
	}
}

func (u *TaskUseCase) SetRepository(repo _interface.ITaskRepository) {
	u.taskRepo = repo
}

func (u *TaskUseCase) CreateTask(req *request.CreateTaskRequest) (*response.CreateTaskResponse, error) {
	// Request ID 생성
	requestID := "task-" + uuid.New().String()

	// Task 모델 생성
	task := &db.Task{
		RequestID:      requestID,
		Status:         db.TaskStatusPending,
		Tasks:          req.Tasks,
		RepositoryName: req.RepositoryName,
		Provider:       req.Provider,
		Interactive:    req.Interactive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Optional 필드 설정
	if req.WorkingDir != "" {
		task.WorkingDir = &req.WorkingDir
	}
	if req.Cmd != "" {
		task.Cmd = &req.Cmd
	}

	// 데이터베이스에 저장
	if err := u.taskRepo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &response.CreateTaskResponse{
		RequestID: requestID,
		Status:    db.TaskStatusPending,
		Message:   "Task created successfully",
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (u *TaskUseCase) ExecuteTask(requestID string) (*response.ExecuteTaskResponse, error) {
	// 태스크 조회
	task, err := u.taskRepo.GetTaskByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// 태스크 상태 확인 (pending 상태만 실행 가능)
	if task.Status != db.TaskStatusPending {
		return nil, fmt.Errorf("task cannot be executed - current status: %s", task.Status)
	}

	// 태스크 상태를 processing으로 변경
	if err := u.taskRepo.UpdateTaskToProcessing(requestID); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// RabbitMQ에 메시지 전송
	taskMessage := &utils.TaskMessage{
		RequestID:      task.RequestID,
		Type:           "task_execution",
		Tasks:          task.Tasks,
		RepositoryName: task.RepositoryName,
		WorkingDir:     task.WorkingDir,
		Cmd:            task.Cmd,
		Provider:       task.Provider,
		Interactive:    task.Interactive,
		Payload: map[string]interface{}{
			"request_type": "task_execution",
			"task_id":      task.ID,
		},
		Timestamp: time.Now(),
	}

	if err := utils.PublishTask(taskMessage); err != nil {
		// RabbitMQ 전송 실패시 상태를 다시 pending으로 롤백
		u.taskRepo.UpdateTaskStatus(requestID, db.TaskStatusPending)
		return nil, fmt.Errorf("failed to queue task for execution: %w", err)
	}

	return &response.ExecuteTaskResponse{
		RequestID: requestID,
		Status:    db.TaskStatusProcessing,
		Message:   "Task queued for execution",
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (u *TaskUseCase) CancelTask(requestID string) (*response.CancelTaskResponse, error) {
	// 태스크 조회
	task, err := u.taskRepo.GetTaskByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// 이미 완료된 태스크는 취소할 수 없음
	if task.Status == db.TaskStatusCompleted || task.Status == db.TaskStatusFailed || task.Status == db.TaskStatusCancelled {
		return nil, fmt.Errorf("task cannot be cancelled - current status: %s", task.Status)
	}

	// 소프트 삭제 (상태를 cancelled로 변경)
	if err := u.taskRepo.CancelTask(requestID); err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return &response.CancelTaskResponse{
		RequestID:   requestID,
		Status:      db.TaskStatusCancelled,
		Message:     "Task cancelled successfully",
		CancelledAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (u *TaskUseCase) ListTasks(req *request.ListTasksRequest) (*response.ListTasksResponse, error) {
	// 기본값 설정
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 태스크 리스트 조회
	tasks, total, err := u.taskRepo.ListTasks(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 응답 모델로 변환
	var taskResponses []response.TaskResponse
	for _, task := range tasks {
		taskResponse := response.TaskResponse{
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
		}
		taskResponses = append(taskResponses, taskResponse)
	}

	// 페이지네이션 정보 계산
	totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)
	hasMore := int64(req.Page) < totalPages

	return &response.ListTasksResponse{
		Tasks:      taskResponses,
		TotalCount: total,
		Page:       req.Page,
		Limit:      req.Limit,
		HasMore:    hasMore,
	}, nil
}

func (u *TaskUseCase) GetTaskStatus(requestID string) (*response.TaskResponse, error) {
	// 태스크 조회
	task, err := u.taskRepo.GetTaskByRequestID(requestID)
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