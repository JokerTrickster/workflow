package repositories

import (
	"context"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
)

// TaskRepository defines the interface for task data persistence
type TaskRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id valueobjects.TaskID) (*entities.Task, error)
	Update(ctx context.Context, task *entities.Task) error
	Delete(ctx context.Context, id valueobjects.TaskID) error
	
	// Query operations
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Task, error)
	GetByStatus(ctx context.Context, status valueobjects.TaskStatus, limit, offset int) ([]*entities.Task, error)
	GetByUserID(ctx context.Context, userID valueobjects.UserID, limit, offset int) ([]*entities.Task, error)
	GetByRepository(ctx context.Context, repository valueobjects.RepositoryPath, limit, offset int) ([]*entities.Task, error)
	GetActiveTasks(ctx context.Context, userID valueobjects.UserID) ([]*entities.Task, error)
	
	// Advanced queries
	GetByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*entities.Task, error)
	GetByEpic(ctx context.Context, epic string, limit, offset int) ([]*entities.Task, error)
	CountByStatus(ctx context.Context, status valueobjects.TaskStatus) (int, error)
	CountByUser(ctx context.Context, userID valueobjects.UserID) (int, error)
	
	// Business-specific queries
	GetTasksRequiringAttention(ctx context.Context) ([]*entities.Task, error) // Failed or long-running tasks
	GetUserTaskStatistics(ctx context.Context, userID valueobjects.UserID) (*TaskStatistics, error)
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(repo TaskRepository) error) error
	
	// Optimistic locking support
	GetByIDWithVersion(ctx context.Context, id valueobjects.TaskID, version int64) (*entities.Task, error)
	UpdateWithVersion(ctx context.Context, task *entities.Task, expectedVersion int64) error
}

// TaskStatistics represents aggregated task statistics for a user
type TaskStatistics struct {
	UserID           valueobjects.UserID
	TotalTasks       int
	CompletedTasks   int
	FailedTasks      int
	PendingTasks     int
	ProcessingTasks  int
	CancelledTasks   int
	TotalTokensUsed  int
	AverageTokensPerTask int
	CompletionRate   float64 // Percentage of completed tasks
	LastActivityAt   *time.Time
}

// TaskFilter represents search/filter criteria for tasks
type TaskFilter struct {
	UserID       *valueobjects.UserID
	Status       *valueobjects.TaskStatus
	Repository   *valueobjects.RepositoryPath
	Epic         *string
	Branch       *valueobjects.BranchName
	CreatedAfter *time.Time
	CreatedBefore *time.Time
	UpdatedAfter *time.Time
	UpdatedBefore *time.Time
	TitleContains *string
	DescriptionContains *string
	MinTokensUsed *int
	MaxTokensUsed *int
}

// TaskOrderBy represents ordering options for task queries
type TaskOrderBy string

const (
	OrderByCreatedAt   TaskOrderBy = "created_at"
	OrderByUpdatedAt   TaskOrderBy = "updated_at"
	OrderByTitle       TaskOrderBy = "title"
	OrderByStatus      TaskOrderBy = "status"
	OrderByTokensUsed  TaskOrderBy = "tokens_used"
)

// TaskOrderDirection represents the direction of ordering
type TaskOrderDirection string

const (
	OrderAsc  TaskOrderDirection = "ASC"
	OrderDesc TaskOrderDirection = "DESC"
)

// QueryOptions represents options for querying tasks
type QueryOptions struct {
	Filter     *TaskFilter
	OrderBy    TaskOrderBy
	Direction  TaskOrderDirection
	Limit      int
	Offset     int
}

// ExtendedTaskRepository provides additional query capabilities
type ExtendedTaskRepository interface {
	TaskRepository
	
	// FindTasks searches for tasks with advanced filtering
	FindTasks(ctx context.Context, options QueryOptions) ([]*entities.Task, int, error)
}