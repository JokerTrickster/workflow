package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/application/container"
	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/delivery/http/routes"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/tests/testutils"
)

func setupAPITestEnvironment(t *testing.T) (*gin.Engine, *testutils.TestEnvironment) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test environment
	env := testutils.SetupTestEnvironment(t)

	// Create application container with test dependencies
	appConfig := &config.AppConfig{
		Server: &config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: &config.DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "testpass",
			Name:     "workflow_test",
			Charset:  "utf8mb4",
		},
		Queue: &config.QueueConfig{
			URL:         "amqp://admin:password@localhost:5672/",
			QueueName:   "task_queue_test",
			Exchange:    "tasks_test",
			RoutingKey:  "task.created",
			MaxRetries:  3,
			RetryDelay:  time.Second,
		},
	}

	// Create container with test repositories
	appContainer := container.NewAppContainer(appConfig)
	
	// Replace repositories with test ones
	appContainer.SetTaskRepository(env.TaskRepo)
	appContainer.SetQueueRepository(env.QueueRepo)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Setup routes
	routes.SetupRoutes(router, appContainer)

	return router, env
}

func TestTaskAPI_CRUD(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	env.CleanDatabase(t)

	t.Run("Create task", func(t *testing.T) {
		createRequest := dto.CreateTaskRequest{
			UserID:      "user123",
			Title:       "API Test Task",
			Description: "Test task via API",
			Repository:  "owner/repo",
			Epic:        "test-epic",
			Branch:      "feature/api-test",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user123")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.CreateTaskResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.TaskID)
		assert.Equal(t, "pending", response.Status)
		assert.NotEmpty(t, response.CreatedAt)

		// Store task ID for subsequent tests
		taskID := response.TaskID

		t.Run("Get task", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s", taskID), nil)
			req.Header.Set("X-User-ID", "user123")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var getResponse dto.GetTaskResponse
			err = json.Unmarshal(w.Body.Bytes(), &getResponse)
			require.NoError(t, err)

			assert.Equal(t, taskID, getResponse.TaskID)
			assert.Equal(t, "user123", getResponse.UserID)
			assert.Equal(t, "API Test Task", getResponse.Title)
			assert.Equal(t, "Test task via API", getResponse.Description)
			assert.Equal(t, "pending", getResponse.Status)
		})

		t.Run("Update task", func(t *testing.T) {
			// First get the task to get current version
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s", taskID), nil)
			req.Header.Set("X-User-ID", "user123")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var getResponse dto.GetTaskResponse
			err = json.Unmarshal(w.Body.Bytes(), &getResponse)
			require.NoError(t, err)

			// Now update the task
			updateRequest := dto.UpdateTaskRequest{
				Title:       testutils.StringPtr("Updated API Test Task"),
				Description: testutils.StringPtr("Updated description via API"),
				Version:     getResponse.Version,
			}

			body, err := json.Marshal(updateRequest)
			require.NoError(t, err)

			req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "user123")

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var updateResponse dto.GetTaskResponse
			err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
			require.NoError(t, err)

			assert.Equal(t, "Updated API Test Task", updateResponse.Title)
			assert.Equal(t, "Updated description via API", updateResponse.Description)
			assert.Greater(t, updateResponse.Version, getResponse.Version)
		})

		t.Run("Cancel task", func(t *testing.T) {
			cancelRequest := dto.TaskActionRequest{
				Reason: "API test cancellation",
			}

			body, err := json.Marshal(cancelRequest)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tasks/%s/cancel", taskID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "user123")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response dto.TaskActionResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, taskID, response.TaskID)
			assert.Equal(t, "cancelled", response.Status)
			assert.Contains(t, response.Message, "cancelled successfully")
		})
	})
}

func TestTaskAPI_ListTasks(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	env.CleanDatabase(t)

	// Create multiple test tasks
	taskIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		createRequest := dto.CreateTaskRequest{
			UserID:      "list-user",
			Title:       fmt.Sprintf("List Test Task %d", i),
			Description: fmt.Sprintf("Test task %d for listing", i),
			Repository:  "owner/repo",
			Epic:        "list-epic",
			Branch:      "feature/list-test",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "list-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response dto.CreateTaskResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		taskIDs[i] = response.TaskID
	}

	t.Run("List all tasks for user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=list-user&limit=10", nil)
		req.Header.Set("X-User-ID", "list-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ListTasksResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Tasks, 5)
		assert.Equal(t, 5, response.Total)
		assert.Equal(t, 10, response.Limit)
		assert.Equal(t, 0, response.Offset)

		// Verify all tasks belong to the user
		for _, task := range response.Tasks {
			assert.Equal(t, "list-user", task.UserID)
		}
	})

	t.Run("List tasks with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=list-user&limit=2&offset=0", nil)
		req.Header.Set("X-User-ID", "list-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ListTasksResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Tasks, 2)
		assert.Equal(t, 5, response.Total)
		assert.Equal(t, 2, response.Limit)
		assert.Equal(t, 0, response.Offset)

		// Test second page
		req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=list-user&limit=2&offset=2", nil)
		req.Header.Set("X-User-ID", "list-user")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Tasks, 2)
		assert.Equal(t, 2, response.Offset)
	})

	t.Run("List tasks with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=list-user&status=pending&repository=owner/repo", nil)
		req.Header.Set("X-User-ID", "list-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ListTasksResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// All tasks should match the filters
		for _, task := range response.Tasks {
			assert.Equal(t, "pending", task.Status)
			assert.Equal(t, "owner/repo", task.Repository)
		}
	})
}

func TestTaskAPI_ErrorCases(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	env.CleanDatabase(t)

	t.Run("Create task with invalid data", func(t *testing.T) {
		createRequest := dto.CreateTaskRequest{
			UserID:      "", // Empty user ID should fail
			Title:       "Test Task",
			Description: "Test description",
			Repository:  "owner/repo",
			Epic:        "test-epic",
			Branch:      "feature/test",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Get non-existent task", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/non-existent-id", nil)
		req.Header.Set("X-User-ID", "user123")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update task with stale version", func(t *testing.T) {
		// First create a task
		createRequest := dto.CreateTaskRequest{
			UserID:      "version-user",
			Title:       "Version Test Task",
			Description: "Test version conflict",
			Repository:  "owner/repo",
			Epic:        "version-epic",
			Branch:      "feature/version-test",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "version-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var createResponse dto.CreateTaskResponse
		err = json.Unmarshal(w.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		// Try to update with a stale version
		updateRequest := dto.UpdateTaskRequest{
			Title:   testutils.StringPtr("Updated Title"),
			Version: 999, // Clearly stale version
		}

		body, err = json.Marshal(updateRequest)
		require.NoError(t, err)

		req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", createResponse.TaskID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "version-user")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Access task without authorization", func(t *testing.T) {
		// Create a task as one user
		createRequest := dto.CreateTaskRequest{
			UserID:      "owner-user",
			Title:       "Private Task",
			Description: "This task should be private",
			Repository:  "owner/repo",
			Epic:        "private-epic",
			Branch:      "feature/private",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "owner-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var createResponse dto.CreateTaskResponse
		err = json.Unmarshal(w.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		// Try to access as different user
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s", createResponse.TaskID), nil)
		req.Header.Set("X-User-ID", "different-user")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestTaskAPI_Statistics(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	env.CleanDatabase(t)

	// Create tasks with different statuses
	userID := "stats-user"
	
	// Create pending tasks
	for i := 0; i < 3; i++ {
		createRequest := dto.CreateTaskRequest{
			UserID:      userID,
			Title:       fmt.Sprintf("Pending Task %d", i),
			Description: "Pending task",
			Repository:  "owner/repo",
			Epic:        "stats-epic",
			Branch:      "feature/stats",
		}

		body, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("Get user statistics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s/statistics", userID), nil)
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskStatisticsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, userID, response.UserID)
		assert.GreaterOrEqual(t, response.TotalTasks, 3)
		assert.GreaterOrEqual(t, response.PendingTasks, 3)
		assert.GreaterOrEqual(t, response.CompletionRate, 0.0)
	})

	t.Run("Get queue statistics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/queue/statistics", nil)
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.QueueStatisticsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, response.TotalEnqueued, 3)
		assert.GreaterOrEqual(t, response.CurrentQueueLength, 0)
		assert.NotEmpty(t, response.LastActivityAt)
	})
}

func TestTaskAPI_HealthCheck(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	t.Run("Health endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "ok", response["status"])
		assert.NotEmpty(t, response["timestamp"])
	})

	t.Run("Ready endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "ready", response["status"])
		assert.Contains(t, response, "database")
		assert.Contains(t, response, "queue")
	})
}

func TestTaskAPI_Concurrent(t *testing.T) {
	router, env := setupAPITestEnvironment(t)
	defer env.TearDown(t)

	env.CleanDatabase(t)

	t.Run("Concurrent task creation", func(t *testing.T) {
		const numWorkers = 10
		const tasksPerWorker = 5
		
		results := make(chan string, numWorkers*tasksPerWorker)
		done := make(chan error, numWorkers)

		for w := 0; w < numWorkers; w++ {
			go func(workerID int) {
				for i := 0; i < tasksPerWorker; i++ {
					createRequest := dto.CreateTaskRequest{
						UserID:      fmt.Sprintf("concurrent-user-%d", workerID),
						Title:       fmt.Sprintf("Concurrent Task %d-%d", workerID, i),
						Description: "Concurrent test task",
						Repository:  "owner/repo",
						Epic:        "concurrent-epic",
						Branch:      "feature/concurrent",
					}

					body, err := json.Marshal(createRequest)
					if err != nil {
						done <- err
						return
					}

					req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("X-User-ID", createRequest.UserID)

					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					if w.Code != http.StatusCreated {
						done <- fmt.Errorf("expected 201, got %d", w.Code)
						return
					}

					var response dto.CreateTaskResponse
					err = json.Unmarshal(w.Body.Bytes(), &response)
					if err != nil {
						done <- err
						return
					}

					results <- response.TaskID
				}
				done <- nil
			}(w)
		}

		// Wait for all workers to complete
		for w := 0; w < numWorkers; w++ {
			err := <-done
			require.NoError(t, err)
		}

		// Collect all task IDs
		close(results)
		taskIDs := make(map[string]bool)
		for taskID := range results {
			taskIDs[taskID] = true
		}

		// Verify we got unique task IDs
		assert.Len(t, taskIDs, numWorkers*tasksPerWorker)
	})
}

// Benchmark tests
func BenchmarkTaskAPI_CreateTask(b *testing.B) {
	router, env := setupAPITestEnvironment(b)
	defer env.TearDown(b)

	env.CleanDatabase(b)

	createRequest := dto.CreateTaskRequest{
		UserID:      "bench-user",
		Title:       "Benchmark Task",
		Description: "Benchmark test task",
		Repository:  "owner/repo",
		Epic:        "bench-epic",
		Branch:      "feature/benchmark",
	}

	body, err := json.Marshal(createRequest)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "bench-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("Expected 201, got %d", w.Code)
		}
	}
}

func BenchmarkTaskAPI_GetTask(b *testing.B) {
	router, env := setupAPITestEnvironment(b)
	defer env.TearDown(b)

	env.CleanDatabase(b)

	// Create a task first
	createRequest := dto.CreateTaskRequest{
		UserID:      "bench-user",
		Title:       "Benchmark Task",
		Description: "Benchmark test task",
		Repository:  "owner/repo",
		Epic:        "bench-epic",
		Branch:      "feature/benchmark",
	}

	body, err := json.Marshal(createRequest)
	if err != nil {
		b.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "bench-user")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		b.Fatalf("Expected 201, got %d", w.Code)
	}

	var response dto.CreateTaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		b.Fatal(err)
	}

	taskID := response.TaskID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s", taskID), nil)
		req.Header.Set("X-User-ID", "bench-user")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}