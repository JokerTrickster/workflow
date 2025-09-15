package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ai-git-workbench/internal/application/container"
	"ai-git-workbench/internal/application/dto"
	"ai-git-workbench/internal/application/interfaces"
)

// TaskUsageExample demonstrates how to use the application layer
func TaskUsageExample() {
	log.Println("🚀 Starting Task Queue Application Layer Example")

	// Initialize application container
	appContainer, err := container.NewApplicationContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application container: %v", err)
	}
	defer appContainer.Close()

	// Check application health
	if err := appContainer.HealthCheck(); err != nil {
		log.Fatalf("❌ Application health check failed: %v", err)
	}
	log.Println("✅ Application health check passed")

	// Get use case
	taskUsecase := appContainer.GetTaskUsecase()
	ctx := context.Background()

	// Example 1: Create a new task
	log.Println("\n📝 Example 1: Creating a new task")
	createTaskExample(ctx, taskUsecase)

	// Example 2: List tasks
	log.Println("\n📋 Example 2: Listing tasks")
	listTasksExample(ctx, taskUsecase)

	// Example 3: Get task statistics
	log.Println("\n📊 Example 3: Getting task statistics")
	getStatisticsExample(ctx, taskUsecase)

	// Example 4: Get queue statistics
	log.Println("\n📈 Example 4: Getting queue statistics")
	getQueueStatisticsExample(ctx, taskUsecase)

	log.Println("\n✅ Task Queue Application Layer Example completed successfully")
}

func createTaskExample(ctx context.Context, usecase interfaces.TaskUsecase) {
	createReq := dto.CreateTaskRequest{
		UserID:      "demo-user-123",
		Title:       "Implement user authentication",
		Description: "Add JWT-based authentication to the API endpoints with proper middleware and validation",
		Repository:  "myorg/backend-service",
		Epic:        "auth-system",
		Branch:      "feature/jwt-auth",
	}

	// Validate request before sending
	if err := createReq.Validate(); err != nil {
		log.Printf("❌ Request validation failed: %v", err)
		return
	}

	response, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		log.Printf("❌ Failed to create task: %v", err)
		return
	}

	log.Printf("✅ Task created successfully:")
	log.Printf("   Task ID: %s", response.TaskID)
	log.Printf("   Status: %s", response.Status)
	log.Printf("   Created At: %s", response.CreatedAt)

	// Store task ID for later examples
	globalTaskID = response.TaskID
}

func listTasksExample(ctx context.Context, usecase interfaces.TaskUsecase) {
	listReq := dto.ListTasksRequest{
		UserID:         stringPtr("demo-user-123"),
		Limit:          10,
		Offset:         0,
		OrderBy:        "created_at",
		OrderDirection: "DESC",
	}

	response, err := usecase.ListTasks(ctx, listReq)
	if err != nil {
		log.Printf("❌ Failed to list tasks: %v", err)
		return
	}

	log.Printf("✅ Found %d tasks (total: %d):", len(response.Tasks), response.Total)
	for i, task := range response.Tasks {
		log.Printf("   %d. %s (ID: %s, Status: %s)", i+1, task.Title, task.TaskID, task.Status)
	}

	if len(response.Tasks) > 0 {
		// Example: Get details of the first task
		getTaskExample(ctx, usecase, response.Tasks[0].TaskID, "demo-user-123")
	}
}

func getTaskExample(ctx context.Context, usecase interfaces.TaskUsecase, taskID, userID string) {
	log.Printf("\n🔍 Getting task details for: %s", taskID)

	response, err := usecase.GetTask(ctx, taskID, userID)
	if err != nil {
		log.Printf("❌ Failed to get task: %v", err)
		return
	}

	// Pretty print the task details
	taskJSON, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("✅ Task details:\n%s", string(taskJSON))

	// Example: Update the task
	updateTaskExample(ctx, usecase, taskID, userID, response.Version)

	// Example: Get task health
	getTaskHealthExample(ctx, usecase, taskID, userID)
}

func updateTaskExample(ctx context.Context, usecase interfaces.TaskUsecase, taskID, userID string, version int64) {
	log.Printf("\n✏️ Updating task: %s", taskID)

	updateReq := dto.UpdateTaskRequest{
		Title:       stringPtr("Implement user authentication with 2FA"),
		Description: stringPtr("Add JWT-based authentication to the API endpoints with proper middleware, validation, and two-factor authentication support"),
		Version:     version,
	}

	response, err := usecase.UpdateTask(ctx, taskID, userID, updateReq)
	if err != nil {
		log.Printf("❌ Failed to update task: %v", err)
		return
	}

	log.Printf("✅ Task updated successfully:")
	log.Printf("   New Title: %s", response.Title)
	log.Printf("   New Version: %d", response.Version)
}

func getTaskHealthExample(ctx context.Context, usecase interfaces.TaskUsecase, taskID, userID string) {
	log.Printf("\n🏥 Checking task health: %s", taskID)

	response, err := usecase.GetTaskHealth(ctx, taskID, userID)
	if err != nil {
		log.Printf("❌ Failed to get task health: %v", err)
		return
	}

	log.Printf("✅ Task health status:")
	log.Printf("   Health: %s", response.Health)
	log.Printf("   Message: %s", response.Message)
	if len(response.Issues) > 0 {
		log.Printf("   Issues: %v", response.Issues)
	}
}

func getStatisticsExample(ctx context.Context, usecase interfaces.TaskUsecase) {
	userID := "demo-user-123"

	response, err := usecase.GetTaskStatistics(ctx, userID)
	if err != nil {
		log.Printf("❌ Failed to get task statistics: %v", err)
		return
	}

	log.Printf("✅ Task statistics for user %s:", userID)
	log.Printf("   Total Tasks: %d", response.TotalTasks)
	log.Printf("   Completed: %d", response.CompletedTasks)
	log.Printf("   Failed: %d", response.FailedTasks)
	log.Printf("   Pending: %d", response.PendingTasks)
	log.Printf("   Processing: %d", response.ProcessingTasks)
	log.Printf("   Cancelled: %d", response.CancelledTasks)
	log.Printf("   Total Tokens Used: %d", response.TotalTokensUsed)
	log.Printf("   Average Tokens Per Task: %d", response.AverageTokensPerTask)
	log.Printf("   Completion Rate: %.2f%%", response.CompletionRate*100)
	if response.LastActivityAt != nil {
		log.Printf("   Last Activity: %s", *response.LastActivityAt)
	}
}

func getQueueStatisticsExample(ctx context.Context, usecase interfaces.TaskUsecase) {
	response, err := usecase.GetQueueStatistics(ctx)
	if err != nil {
		log.Printf("❌ Failed to get queue statistics: %v", err)
		return
	}

	log.Printf("✅ Queue statistics:")
	log.Printf("   Total Enqueued: %d", response.TotalEnqueued)
	log.Printf("   Total Dequeued: %d", response.TotalDequeued)
	log.Printf("   Total Processed: %d", response.TotalProcessed)
	log.Printf("   Total Failed: %d", response.TotalFailed)
	log.Printf("   Current Queue Length: %d", response.CurrentQueueLength)
	log.Printf("   Average Processing Time: %s", response.AverageProcessingTime)
	log.Printf("   Throughput Per Minute: %.2f", response.ThroughputPerMinute)
	log.Printf("   Error Rate: %.2f%%", response.ErrorRate*100)
	log.Printf("   Dead Letter Count: %d", response.DeadLetterCount)
	log.Printf("   Workers Active: %d", response.WorkersActive)
	log.Printf("   Last Activity: %s", response.LastActivityAt)
}

// TaskLifecycleExample demonstrates the complete task lifecycle
func TaskLifecycleExample() {
	log.Println("🔄 Starting Task Lifecycle Example")

	// Initialize application container
	appContainer, err := container.NewApplicationContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application container: %v", err)
	}
	defer appContainer.Close()

	usecase := appContainer.GetTaskUsecase()
	ctx := context.Background()
	userID := "lifecycle-demo-user"

	// Step 1: Create a task
	log.Println("\n1. Creating task...")
	createReq := dto.CreateTaskRequest{
		UserID:      userID,
		Title:       "Lifecycle demo task",
		Description: "This task demonstrates the complete lifecycle",
		Repository:  "demo/lifecycle",
		Epic:        "demo-epic",
		Branch:      "feature/lifecycle-demo",
	}

	createResp, err := usecase.CreateTask(ctx, createReq)
	if err != nil {
		log.Fatalf("❌ Failed to create task: %v", err)
	}
	log.Printf("✅ Task created: %s (Status: %s)", createResp.TaskID, createResp.Status)

	taskID := createResp.TaskID

	// Step 2: Get task details
	log.Println("\n2. Getting task details...")
	task, err := usecase.GetTask(ctx, taskID, userID)
	if err != nil {
		log.Fatalf("❌ Failed to get task: %v", err)
	}
	log.Printf("✅ Task retrieved: %s (Status: %s)", task.Title, task.Status)

	// Step 3: Update task
	log.Println("\n3. Updating task...")
	updateReq := dto.UpdateTaskRequest{
		Description: stringPtr("Updated description with more details"),
		Version:     task.Version,
	}

	updatedTask, err := usecase.UpdateTask(ctx, taskID, userID, updateReq)
	if err != nil {
		log.Fatalf("❌ Failed to update task: %v", err)
	}
	log.Printf("✅ Task updated: Version %d -> %d", task.Version, updatedTask.Version)

	// Step 4: Cancel task
	log.Println("\n4. Cancelling task...")
	cancelReq := dto.TaskActionRequest{
		Reason: "Demo cancellation for lifecycle example",
	}

	cancelResp, err := usecase.CancelTask(ctx, taskID, userID, cancelReq)
	if err != nil {
		log.Fatalf("❌ Failed to cancel task: %v", err)
	}
	log.Printf("✅ Task cancelled: %s (Status: %s)", cancelResp.TaskID, cancelResp.Status)

	// Step 5: Resume task
	log.Println("\n5. Resuming task...")
	resumeReq := dto.TaskActionRequest{
		Reason: "Demo resume for lifecycle example",
	}

	resumeResp, err := usecase.ResumeTask(ctx, taskID, userID, resumeReq)
	if err != nil {
		log.Fatalf("❌ Failed to resume task: %v", err)
	}
	log.Printf("✅ Task resumed: %s (Status: %s)", resumeResp.TaskID, resumeResp.Status)

	// Step 6: Get final statistics
	log.Println("\n6. Getting final statistics...")
	stats, err := usecase.GetTaskStatistics(ctx, userID)
	if err != nil {
		log.Fatalf("❌ Failed to get statistics: %v", err)
	}
	log.Printf("✅ Final statistics: %d total tasks, %d pending", stats.TotalTasks, stats.PendingTasks)

	log.Println("\n✅ Task Lifecycle Example completed successfully")
}

// ErrorHandlingExample demonstrates error handling patterns
func ErrorHandlingExample() {
	log.Println("⚠️ Starting Error Handling Example")

	appContainer, err := container.NewApplicationContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application container: %v", err)
	}
	defer appContainer.Close()

	usecase := appContainer.GetTaskUsecase()
	ctx := context.Background()

	// Example 1: Validation errors
	log.Println("\n1. Testing validation errors...")
	invalidReq := dto.CreateTaskRequest{
		UserID:      "", // Invalid: empty user ID
		Title:       "",  // Invalid: empty title
		Repository:  "invalid-repo-format", // Invalid: wrong format
		Epic:        "",  // Invalid: empty epic
		Branch:      "",  // Invalid: empty branch
	}

	_, err = usecase.CreateTask(ctx, invalidReq)
	if err != nil {
		log.Printf("✅ Expected validation error: %v", err)
	} else {
		log.Printf("❌ Expected validation error but got success")
	}

	// Example 2: Authorization errors
	log.Println("\n2. Testing authorization errors...")
	// First create a task with one user
	validReq := dto.CreateTaskRequest{
		UserID:      "user1",
		Title:       "User1's task",
		Description: "This task belongs to user1",
		Repository:  "user1/repo",
		Epic:        "user1-epic",
		Branch:      "feature/test",
	}

	createResp, err := usecase.CreateTask(ctx, validReq)
	if err != nil {
		log.Printf("❌ Failed to create task for authorization test: %v", err)
		return
	}

	// Try to access with different user
	_, err = usecase.GetTask(ctx, createResp.TaskID, "user2")
	if err != nil {
		log.Printf("✅ Expected authorization error: %v", err)
	} else {
		log.Printf("❌ Expected authorization error but got success")
	}

	// Example 3: Not found errors
	log.Println("\n3. Testing not found errors...")
	_, err = usecase.GetTask(ctx, "non-existent-task-id", "user1")
	if err != nil {
		log.Printf("✅ Expected not found error: %v", err)
	} else {
		log.Printf("❌ Expected not found error but got success")
	}

	log.Println("\n✅ Error Handling Example completed")
}

// Global variable to store task ID for examples
var globalTaskID string

// Helper function
func stringPtr(s string) *string {
	return &s
}