package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/application/interfaces"
	"ai-git-workbench/internal/application/usecases"
	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/services"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/tests/testutils"
)

// Mock implementations
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id valueobjects.TaskID) (*entities.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id valueobjects.TaskID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByUserID(ctx context.Context, userID valueobjects.UserID, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByStatus(ctx context.Context, status valueobjects.TaskStatus, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *MockTaskRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *MockTaskRepository) WithTransaction(ctx context.Context, fn func(repositories.TaskRepository) error) error {
	args := m.Called(ctx, mock.AnythingOfType("func(repositories.TaskRepository) error"))
	return args.Error(0)
}

func (m *MockTaskRepository) GetUserTaskStatistics(ctx context.Context, userID valueobjects.UserID) (*repositories.UserTaskStatistics, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*repositories.UserTaskStatistics), args.Error(1)
}

func (m *MockTaskRepository) GetTasksRequiringAttention(ctx context.Context) ([]*entities.Task, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Task), args.Error(1)
}

type MockQueueRepository struct {
	mock.Mock
}

func (m *MockQueueRepository) Enqueue(ctx context.Context, task *entities.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockQueueRepository) Dequeue(ctx context.Context) (*entities.Task, error) {
	args := m.Called(ctx)
	return args.Get(0).(*entities.Task), args.Error(1)
}

func (m *MockQueueRepository) RemoveTask(ctx context.Context, taskID valueobjects.TaskID) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockQueueRepository) GetQueueLength(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockQueueRepository) GetQueueStatistics(ctx context.Context) (*repositories.QueueStatistics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*repositories.QueueStatistics), args.Error(1)
}

func (m *MockQueueRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CanAccessTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockAuthorizationService) CanModifyTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockAuthorizationService) CanDeleteTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockAuthorizationService) CanViewStatistics(ctx context.Context, requestingUserID valueobjects.UserID, targetUserID *valueobjects.UserID) error {
	args := m.Called(ctx, requestingUserID, targetUserID)
	return args.Error(0)
}

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) PublishTaskCreated(ctx context.Context, task *dto.GetTaskResponse) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEventService) PublishTaskUpdated(ctx context.Context, task *dto.GetTaskResponse) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEventService) PublishTaskDeleted(ctx context.Context, taskID string, userID string) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func (m *MockEventService) PublishTaskStatusChanged(ctx context.Context, taskID string, oldStatus string, newStatus string) error {
	args := m.Called(ctx, taskID, oldStatus, newStatus)
	return args.Error(0)
}

func (m *MockEventService) PublishTaskResumed(ctx context.Context, taskID string, reason string) error {
	args := m.Called(ctx, taskID, reason)
	return args.Error(0)
}

// Test setup helpers
func setupTaskUsecase(t *testing.T) (
	interfaces.TaskUsecase,
	*MockTaskRepository,
	*MockQueueRepository,
	*MockAuthorizationService,
	*MockEventService,
) {
	mockTaskRepo := &MockTaskRepository{}
	mockQueueRepo := &MockQueueRepository{}
	mockAuthService := &MockAuthorizationService{}
	mockEventService := &MockEventService{}

	validationService := services.NewTaskValidationService()
	lifecycleService := services.NewTaskLifecycleService(mockTaskRepo, mockQueueRepo)

	usecase := usecases.NewTaskUsecase(
		mockTaskRepo,
		mockQueueRepo,
		validationService,
		lifecycleService,
		mockAuthService,
		mockEventService,
	)

	return usecase, mockTaskRepo, mockQueueRepo, mockAuthService, mockEventService
}

func TestTaskUsecase_CreateTask(t *testing.T) {
	usecase, mockTaskRepo, mockQueueRepo, mockAuthService, mockEventService := setupTaskUsecase(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		request     dto.CreateTaskRequest
		setupMocks  func()
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid task creation",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "Test description",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			setupMocks: func() {
				mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entities.Task")).Return(nil)
				mockQueueRepo.On("Enqueue", ctx, mock.AnythingOfType("*entities.Task")).Return(nil)
				mockEventService.On("PublishTaskCreated", ctx, mock.AnythingOfType("*dto.GetTaskResponse")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Invalid request validation should fail",
			request: dto.CreateTaskRequest{
				UserID:      "",
				Title:       "Test Task",
				Description: "Test description",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			setupMocks: func() {
				// No mocks needed for validation failure
			},
			wantErr:     true,
			errContains: "user_id is required",
		},
		{
			name: "Repository creation failure should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "Test description",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			setupMocks: func() {
				mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entities.Task")).Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name: "Queue enqueue failure should fail",
			request: dto.CreateTaskRequest{
				UserID:      "user123",
				Title:       "Test Task",
				Description: "Test description",
				Repository:  "owner/repo",
				Epic:        "epic-1",
				Branch:      "feature/test",
			},
			setupMocks: func() {
				mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entities.Task")).Return(nil)
				mockQueueRepo.On("Enqueue", ctx, mock.AnythingOfType("*entities.Task")).Return(errors.New("queue error"))
			},
			wantErr:     true,
			errContains: "queue error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockTaskRepo.ExpectedCalls = nil
			mockQueueRepo.ExpectedCalls = nil
			mockEventService.ExpectedCalls = nil

			tt.setupMocks()

			response, err := usecase.CreateTask(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.TaskID)
				assert.Equal(t, "pending", response.Status)
				assert.NotEmpty(t, response.CreatedAt)
			}

			mockTaskRepo.AssertExpectations(t)
			mockQueueRepo.AssertExpectations(t)
			mockEventService.AssertExpectations(t)
		})
	}
}

func TestTaskUsecase_GetTask(t *testing.T) {
	usecase, mockTaskRepo, _, mockAuthService, _ := setupTaskUsecase(t)
	ctx := context.Background()

	// Create a test task for retrieval
	testTask, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		userID      string
		setupMocks  func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "Valid task retrieval",
			taskID: testTask.ID().Value(),
			userID: "user123",
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user123")
				mockAuthService.On("CanAccessTask", ctx, userID, taskID).Return(nil)
				mockTaskRepo.On("GetByID", ctx, taskID).Return(testTask, nil)
			},
			wantErr: false,
		},
		{
			name:   "Invalid task ID should fail",
			taskID: "invalid-id",
			userID: "user123",
			setupMocks: func() {
				// No mocks needed for invalid ID
			},
			wantErr:     true,
			errContains: "invalid task ID",
		},
		{
			name:   "Authorization failure should fail",
			taskID: testTask.ID().Value(),
			userID: "user456",
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user456")
				mockAuthService.On("CanAccessTask", ctx, userID, taskID).Return(errors.New("access denied"))
			},
			wantErr:     true,
			errContains: "access denied",
		},
		{
			name:   "Task not found should fail",
			taskID: testTask.ID().Value(),
			userID: "user123",
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user123")
				mockAuthService.On("CanAccessTask", ctx, userID, taskID).Return(nil)
				mockTaskRepo.On("GetByID", ctx, taskID).Return((*entities.Task)(nil), errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockTaskRepo.ExpectedCalls = nil
			mockAuthService.ExpectedCalls = nil

			tt.setupMocks()

			response, err := usecase.GetTask(ctx, tt.taskID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, testTask.ID().Value(), response.TaskID)
				assert.Equal(t, testTask.Title(), response.Title)
				assert.Equal(t, testTask.Status().Value(), response.Status)
			}

			mockTaskRepo.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestTaskUsecase_UpdateTask(t *testing.T) {
	usecase, mockTaskRepo, _, mockAuthService, mockEventService := setupTaskUsecase(t)
	ctx := context.Background()

	// Create a test task for updating
	testTask, err := testutils.CreateTestTask("user123", "Original Title")
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		userID      string
		request     dto.UpdateTaskRequest
		setupMocks  func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "Valid task update",
			taskID: testTask.ID().Value(),
			userID: "user123",
			request: dto.UpdateTaskRequest{
				Title:   testutils.StringPtr("Updated Title"),
				Version: testTask.Version(),
			},
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user123")
				
				mockAuthService.On("CanModifyTask", ctx, userID, taskID).Return(nil)
				mockTaskRepo.On("GetByID", ctx, taskID).Return(testTask, nil)
				mockTaskRepo.On("WithTransaction", ctx, mock.AnythingOfType("func(repositories.TaskRepository) error")).Return(nil)
				mockEventService.On("PublishTaskUpdated", ctx, mock.AnythingOfType("*dto.GetTaskResponse")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Version conflict should fail",
			taskID: testTask.ID().Value(),
			userID: "user123",
			request: dto.UpdateTaskRequest{
				Title:   testutils.StringPtr("Updated Title"),
				Version: testTask.Version() - 1, // Stale version
			},
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user123")
				
				mockAuthService.On("CanModifyTask", ctx, userID, taskID).Return(nil)
				mockTaskRepo.On("GetByID", ctx, taskID).Return(testTask, nil)
			},
			wantErr:     true,
			errContains: "modified by another process",
		},
		{
			name:   "Authorization failure should fail",
			taskID: testTask.ID().Value(),
			userID: "user456",
			request: dto.UpdateTaskRequest{
				Title:   testutils.StringPtr("Updated Title"),
				Version: testTask.Version(),
			},
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user456")
				
				mockAuthService.On("CanModifyTask", ctx, userID, taskID).Return(errors.New("access denied"))
			},
			wantErr:     true,
			errContains: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockTaskRepo.ExpectedCalls = nil
			mockAuthService.ExpectedCalls = nil
			mockEventService.ExpectedCalls = nil

			tt.setupMocks()

			response, err := usecase.UpdateTask(ctx, tt.taskID, tt.userID, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				if tt.request.Title != nil {
					assert.Equal(t, *tt.request.Title, response.Title)
				}
				assert.Greater(t, response.Version, tt.request.Version)
			}

			mockTaskRepo.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
			mockEventService.AssertExpectations(t)
		})
	}
}

func TestTaskUsecase_CancelTask(t *testing.T) {
	usecase, mockTaskRepo, mockQueueRepo, mockAuthService, mockEventService := setupTaskUsecase(t)
	ctx := context.Background()

	// Create a test task for cancellation
	testTask, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		userID      string
		request     dto.TaskActionRequest
		setupMocks  func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "Valid task cancellation",
			taskID: testTask.ID().Value(),
			userID: "user123",
			request: dto.TaskActionRequest{
				Reason: "User requested cancellation",
			},
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user123")
				
				mockAuthService.On("CanModifyTask", ctx, userID, taskID).Return(nil)
				mockTaskRepo.On("GetByID", ctx, taskID).Return(testTask, nil)
				mockTaskRepo.On("WithTransaction", ctx, mock.AnythingOfType("func(repositories.TaskRepository) error")).Return(nil)
				mockQueueRepo.On("RemoveTask", ctx, taskID).Return(nil)
				mockEventService.On("PublishTaskStatusChanged", ctx, testTask.ID().Value(), "pending", "cancelled").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Authorization failure should fail",
			taskID: testTask.ID().Value(),
			userID: "user456",
			request: dto.TaskActionRequest{
				Reason: "Unauthorized cancellation",
			},
			setupMocks: func() {
				taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
				userID, _ := valueobjects.NewUserID("user456")
				
				mockAuthService.On("CanModifyTask", ctx, userID, taskID).Return(errors.New("access denied"))
			},
			wantErr:     true,
			errContains: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockTaskRepo.ExpectedCalls = nil
			mockQueueRepo.ExpectedCalls = nil
			mockAuthService.ExpectedCalls = nil
			mockEventService.ExpectedCalls = nil

			tt.setupMocks()

			response, err := usecase.CancelTask(ctx, tt.taskID, tt.userID, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.taskID, response.TaskID)
				assert.Equal(t, "cancelled", response.Status)
				assert.Contains(t, response.Message, "cancelled successfully")
			}

			mockTaskRepo.AssertExpectations(t)
			mockQueueRepo.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
			mockEventService.AssertExpectations(t)
		})
	}
}

func TestTaskUsecase_GetTaskStatistics(t *testing.T) {
	usecase, mockTaskRepo, _, mockAuthService, _ := setupTaskUsecase(t)
	ctx := context.Background()

	userID := "user123"
	userIDVO, _ := valueobjects.NewUserID(userID)

	stats := &repositories.UserTaskStatistics{
		UserID:               userIDVO,
		TotalTasks:           10,
		CompletedTasks:       6,
		FailedTasks:          2,
		PendingTasks:         1,
		ProcessingTasks:      1,
		CancelledTasks:       0,
		TotalTokensUsed:      1000,
		AverageTokensPerTask: 100.0,
		CompletionRate:       0.6,
		LastActivityAt:       testutils.TimePtr(time.Now()),
	}

	mockAuthService.On("CanViewStatistics", ctx, userIDVO, &userIDVO).Return(nil)
	mockTaskRepo.On("GetUserTaskStatistics", ctx, userIDVO).Return(stats, nil)

	response, err := usecase.GetTaskStatistics(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, 10, response.TotalTasks)
	assert.Equal(t, 6, response.CompletedTasks)
	assert.Equal(t, 2, response.FailedTasks)
	assert.Equal(t, 1, response.PendingTasks)
	assert.Equal(t, 1, response.ProcessingTasks)
	assert.Equal(t, 0, response.CancelledTasks)
	assert.Equal(t, 1000, response.TotalTokensUsed)
	assert.Equal(t, 100.0, response.AverageTokensPerTask)
	assert.Equal(t, 0.6, response.CompletionRate)
	assert.NotNil(t, response.LastActivityAt)

	mockAuthService.AssertExpectations(t)
	mockTaskRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkTaskUsecase_CreateTask(b *testing.B) {
	usecase, mockTaskRepo, mockQueueRepo, _, mockEventService := setupTaskUsecase(b)
	ctx := context.Background()

	request := dto.CreateTaskRequest{
		UserID:      "user123",
		Title:       "Benchmark Task",
		Description: "Benchmark description",
		Repository:  "owner/repo",
		Epic:        "epic-1",
		Branch:      "feature/benchmark",
	}

	// Setup mocks for all iterations
	mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entities.Task")).Return(nil)
	mockQueueRepo.On("Enqueue", ctx, mock.AnythingOfType("*entities.Task")).Return(nil)
	mockEventService.On("PublishTaskCreated", ctx, mock.AnythingOfType("*dto.GetTaskResponse")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := usecase.CreateTask(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTaskUsecase_GetTask(b *testing.B) {
	usecase, mockTaskRepo, _, mockAuthService, _ := setupTaskUsecase(b)
	ctx := context.Background()

	testTask, err := testutils.CreateTestTask("user123", "Benchmark Task")
	if err != nil {
		b.Fatal(err)
	}

	taskID, _ := valueobjects.ParseTaskID(testTask.ID().Value())
	userID, _ := valueobjects.NewUserID("user123")

	// Setup mocks for all iterations
	mockAuthService.On("CanAccessTask", ctx, userID, taskID).Return(nil)
	mockTaskRepo.On("GetByID", ctx, taskID).Return(testTask, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := usecase.GetTask(ctx, testTask.ID().Value(), "user123")
		if err != nil {
			b.Fatal(err)
		}
	}
}