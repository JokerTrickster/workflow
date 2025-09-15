package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/internal/infrastructure/database"
	mysqlRepo "ai-git-workbench/internal/infrastructure/repositories"
	"ai-git-workbench/internal/infrastructure/queue"
)

// TestEnvironment encapsulates test environment resources
type TestEnvironment struct {
	DB             *database.DB
	TaskRepo       repositories.TaskRepository
	QueueRepo      repositories.QueueRepository
	MySQLContainer *mysql.MySQLContainer
	RabbitContainer *rabbitmq.RabbitMQContainer
	Context        context.Context
}

// SetupTestEnvironment creates a complete test environment with containers
func SetupTestEnvironment(t testing.TB) *TestEnvironment {
	ctx := context.Background()

	// Start MySQL container
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("workflow_test"),
		mysql.WithUsername("root"),
		mysql.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306").
				WithOccurrence(1).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	// Start RabbitMQ container
	rabbitContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12-management",
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("completed with").
				WithOccurrence(1).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	// Get MySQL connection details
	host, err := mysqlContainer.Host(ctx)
	require.NoError(t, err)
	
	port, err := mysqlContainer.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// Create database connection
	dbConfig := &config.DatabaseConfig{
		Host:     host,
		Port:     port.Port(),
		User:     "root",
		Password: "testpass",
		Name:     "workflow_test",
		Charset:  "utf8mb4",
	}

	db, err := database.NewMySQLConnection(dbConfig)
	require.NoError(t, err)

	// Run migrations
	err = runMigrations(db)
	require.NoError(t, err)

	// Create repositories
	taskRepo := mysqlRepo.NewMySQLTaskRepository(db)

	// Get RabbitMQ connection details
	rabbitHost, err := rabbitContainer.Host(ctx)
	require.NoError(t, err)
	
	rabbitPort, err := rabbitContainer.MappedPort(ctx, "5672")
	require.NoError(t, err)

	// Create queue repository
	queueConfig := &config.QueueConfig{
		URL:         fmt.Sprintf("amqp://admin:password@%s:%s/", rabbitHost, rabbitPort.Port()),
		QueueName:   "task_queue_test",
		Exchange:    "tasks_test",
		RoutingKey:  "task.created",
		MaxRetries:  3,
		RetryDelay:  time.Second,
	}

	queueRepo, err := queue.NewRabbitMQRepository(queueConfig)
	require.NoError(t, err)

	return &TestEnvironment{
		DB:              db,
		TaskRepo:        taskRepo,
		QueueRepo:       queueRepo,
		MySQLContainer:  mysqlContainer,
		RabbitContainer: rabbitContainer,
		Context:         ctx,
	}
}

// TearDown cleans up the test environment
func (env *TestEnvironment) TearDown(t testing.TB) {
	if env.DB != nil {
		env.DB.Close()
	}
	if env.MySQLContainer != nil {
		require.NoError(t, env.MySQLContainer.Terminate(env.Context))
	}
	if env.RabbitContainer != nil {
		require.NoError(t, env.RabbitContainer.Terminate(env.Context))
	}
}

// CleanDatabase truncates all tables for a clean test state
func (env *TestEnvironment) CleanDatabase(t testing.TB) {
	_, err := env.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
	require.NoError(t, err)
	
	_, err = env.DB.Exec("TRUNCATE TABLE task_metadata")
	require.NoError(t, err)
	
	_, err = env.DB.Exec("TRUNCATE TABLE tasks")
	require.NoError(t, err)
	
	_, err = env.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	require.NoError(t, err)
}

// CreateTestTask creates a test task with default values
func CreateTestTask(userID, title string) (*entities.Task, error) {
	user, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, err
	}
	
	repo, err := valueobjects.NewRepositoryPath("test/repo")
	if err != nil {
		return nil, err
	}
	
	branch, err := valueobjects.NewBranchName("feature/test")
	if err != nil {
		return nil, err
	}

	return entities.CreateTask(
		user,
		title,
		"Test task description",
		repo,
		"test-epic",
		branch,
	)
}

// CreateTestTaskWithDetails creates a test task with specific details
func CreateTestTaskWithDetails(
	userID, title, description, repository, epic, branch string,
) (*entities.Task, error) {
	user, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, err
	}
	
	repo, err := valueobjects.NewRepositoryPath(repository)
	if err != nil {
		return nil, err
	}
	
	branchVO, err := valueobjects.NewBranchName(branch)
	if err != nil {
		return nil, err
	}

	return entities.CreateTask(user, title, description, repo, epic, branchVO)
}

// AssertTaskEqual asserts that two tasks have equal essential properties
func AssertTaskEqual(t testing.TB, expected, actual *entities.Task) {
	require.Equal(t, expected.ID().Value(), actual.ID().Value())
	require.Equal(t, expected.UserID().Value(), actual.UserID().Value())
	require.Equal(t, expected.Title(), actual.Title())
	require.Equal(t, expected.Description(), actual.Description())
	require.Equal(t, expected.Status().Value(), actual.Status().Value())
	require.Equal(t, expected.Repository().Value(), actual.Repository().Value())
	require.Equal(t, expected.Epic(), actual.Epic())
	require.Equal(t, expected.Branch().Value(), actual.Branch().Value())
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t testing.TB, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	timeoutCh := time.After(timeout)
	
	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
	}
}

// runMigrations runs database migrations for testing
func runMigrations(db *database.DB) error {
	// Create tasks table
	tasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		title VARCHAR(500) NOT NULL,
		description TEXT,
		status VARCHAR(50) NOT NULL,
		repository VARCHAR(255) NOT NULL,
		epic VARCHAR(255) NOT NULL,
		branch VARCHAR(255) NOT NULL,
		tokens_used INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		started_at TIMESTAMP NULL,
		completed_at TIMESTAMP NULL,
		version BIGINT DEFAULT 1,
		INDEX idx_user_id (user_id),
		INDEX idx_status (status),
		INDEX idx_repository (repository),
		INDEX idx_epic (epic),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	if _, err := db.Exec(tasksTable); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create task_metadata table
	metadataTable := `
	CREATE TABLE IF NOT EXISTS task_metadata (
		task_id VARCHAR(36) NOT NULL,
		metadata_key VARCHAR(255) NOT NULL,
		metadata_value TEXT,
		PRIMARY KEY (task_id, metadata_key),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	if _, err := db.Exec(metadataTable); err != nil {
		return fmt.Errorf("failed to create task_metadata table: %w", err)
	}

	return nil
}

// BenchmarkConfig provides configuration for performance tests
type BenchmarkConfig struct {
	ConcurrentUsers    int
	TasksPerUser      int
	MaxResponseTime   time.Duration
	ExpectedThroughput int // operations per second
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		ConcurrentUsers:    10,
		TasksPerUser:      10,
		MaxResponseTime:   200 * time.Millisecond,
		ExpectedThroughput: 100,
	}
}

// MockTime provides a controllable time source for testing
type MockTime struct {
	currentTime time.Time
}

// NewMockTime creates a new mock time starting at the given time
func NewMockTime(startTime time.Time) *MockTime {
	return &MockTime{currentTime: startTime}
}

// Now returns the current mock time
func (m *MockTime) Now() time.Time {
	return m.currentTime
}

// Advance moves the mock time forward by the given duration
func (m *MockTime) Advance(duration time.Duration) {
	m.currentTime = m.currentTime.Add(duration)
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns a pointer to the given time
func TimePtr(t time.Time) *time.Time {
	return &t
}