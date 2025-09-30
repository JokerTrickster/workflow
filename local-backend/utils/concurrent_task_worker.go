package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/features/github/model/response"
	"main/features/github/service"
	"main/utils/db/mysql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// ConcurrentTaskWorker handles multiple RabbitMQ messages concurrently
type ConcurrentTaskWorker struct {
	connection      *amqp.Connection
	channel         *amqp.Channel
	queueName       string
	rabbitMQURL     string
	providerFactory *AIProviderFactory
	failureCount    map[string]int
	lastFailureTime map[string]time.Time
	maxRetries      int
	db              *gorm.DB

	// Concurrency control
	maxConcurrent   int
	semaphore       chan struct{}
	wg              sync.WaitGroup
	mutex           sync.RWMutex // Protect shared state

	// GitHub integration
	githubService   *service.GitHubService
	githubToken     string

	// Branch management
	branchManager   *BranchManager

	// Error handling
	errorHandler    *ErrorHandler
}

// NewConcurrentTaskWorker creates a new concurrent task worker
func NewConcurrentTaskWorker(rabbitMQURL, queueName string, maxConcurrent int) *ConcurrentTaskWorker {
	if maxConcurrent <= 0 {
		maxConcurrent = 3 // Default to 3 concurrent tasks
	}

	// Safely get database connection
	var db *gorm.DB
	if mysql.GormMysqlDB != nil {
		db = mysql.GormMysqlDB
		log.Printf("ConcurrentTaskWorker: Database connection available")
	} else {
		log.Printf("ConcurrentTaskWorker: Database connection not available, failure logging will be disabled")
	}

	return &ConcurrentTaskWorker{
		rabbitMQURL:     rabbitMQURL,
		queueName:       queueName,
		providerFactory: GlobalAIProviderFactory,
		failureCount:    make(map[string]int),
		lastFailureTime: make(map[string]time.Time),
		maxRetries:      5,
		db:              db,
		maxConcurrent:   maxConcurrent,
		semaphore:       make(chan struct{}, maxConcurrent),
		githubService:   service.NewGitHubService(),
		githubToken:     os.Getenv("GITHUB_TOKEN"),
		branchManager:   GetGlobalBranchManager(),
		errorHandler:    GetGlobalErrorHandler(),
	}
}

// Connect establishes connection to RabbitMQ
func (w *ConcurrentTaskWorker) Connect() error {
	conn, err := amqp.Dial(w.rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue to ensure it exists
	_, err = ch.QueueDeclare(
		w.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	w.connection = conn
	w.channel = ch

	return nil
}

// StartConsuming starts consuming messages with concurrency control
func (w *ConcurrentTaskWorker) StartConsuming(ctx context.Context) error {
	if w.channel == nil {
		return fmt.Errorf("worker not connected")
	}

	// Set QoS to allow multiple unacknowledged messages
	err := w.channel.Qos(w.maxConcurrent*2, 0, false) // Prefetch more for better throughput
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := w.channel.Consume(
		w.queueName, // queue
		"",          // consumer
		false,       // auto-ack (manual ack for reliability)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Concurrent task worker started with %d max concurrent tasks, waiting for messages...", w.maxConcurrent)

	for {
		select {
		case <-ctx.Done():
			log.Println("Concurrent task worker context cancelled, waiting for running tasks to complete...")
			w.wg.Wait() // Wait for all running tasks to complete
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				w.wg.Wait() // Wait for all running tasks to complete
				return fmt.Errorf("message channel closed")
			}

			// Start task in goroutine with concurrency control
			w.wg.Add(1)
			go w.handleMessageConcurrent(ctx, msg)
		}
	}
}

// handleMessageConcurrent processes messages with semaphore-based concurrency control
func (w *ConcurrentTaskWorker) handleMessageConcurrent(ctx context.Context, msg amqp.Delivery) {
	defer w.wg.Done()

	// Acquire semaphore (blocks if at max capacity)
	select {
	case w.semaphore <- struct{}{}:
		defer func() { <-w.semaphore }() // Release semaphore when done
	case <-ctx.Done():
		msg.Nack(false, true) // Requeue if context cancelled
		return
	}

	log.Printf("Processing message (concurrent): %s", string(msg.Body))

	// Parse message
	var taskMsg TaskMessage
	if err := json.Unmarshal(msg.Body, &taskMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	// Update task status to "processing" in database before starting execution
	if err := w.updateTaskStatusInDB(&taskMsg, "processing"); err != nil {
		log.Printf("Warning: Failed to update task status to processing: %v", err)
		// Continue with task execution even if DB update fails
	}

	// Create GitHub issue before task execution
	issueURL, issueNumber, err := w.createGitHubIssue(ctx, &taskMsg)
	if err != nil {
		log.Printf("Warning: Failed to create GitHub issue: %v", err)
		// Continue with task execution even if issue creation fails
	}

	// Create task branch with conflict prevention
	branchInfo, err := w.branchManager.CreateTaskBranch(ctx, &taskMsg)
	if err != nil {
		log.Printf("Warning: Failed to create task branch: %v", err)
		// Continue with task execution on current branch
	}

	// Execute task
	result, err := w.executeTaskConcurrent(ctx, &taskMsg)
	if err != nil {
		log.Printf("Failed to execute task: %v", err)

		// Handle error with comprehensive error handling
		taskErr := w.errorHandler.ClassifyError(err, &taskMsg)
		if handleErr := w.errorHandler.HandleTaskError(ctx, taskErr); handleErr != nil {
			log.Printf("Error handling failed: %v", handleErr)
		}

		w.recordTaskFailure(&taskMsg, err.Error())

		// Update task status to "failed" in database
		if updateErr := w.updateTaskStatusInDB(&taskMsg, "failed"); updateErr != nil {
			log.Printf("Warning: Failed to update task status to failed: %v", updateErr)
		}

		// Mark branch as completed (failed)
		if branchInfo != nil {
			w.branchManager.CompleteBranch(taskMsg.RepositoryName, false)
		}

		msg.Ack(false)
		return
	}

	// Record successful task completion to database
	w.recordTaskSuccess(&taskMsg, result, issueURL)

	// Mark branch as completed
	if branchInfo != nil {
		w.branchManager.CompleteBranch(taskMsg.RepositoryName, true)
	}

	// Handle successful completion - commit and push changes first
	if w.shouldCommitChanges(result) {
		if err := w.commitAndPushChanges(ctx, &taskMsg, result); err != nil {
			log.Printf("Failed to commit and push changes: %v", err)

			// Handle Git operation errors
			taskErr := w.errorHandler.ClassifyError(err, &taskMsg)
			if taskErr.Type == ErrorTypeGit && taskErr.Recoverable {
				if handleErr := w.errorHandler.HandleTaskError(ctx, taskErr); handleErr != nil {
					log.Printf("Git error handling failed: %v", handleErr)
				}
			}
			// Continue with PR creation even if commit/push fails
		}
	}

	// Check if PR needed after committing changes
	if w.shouldCreatePR(result) {
		if err := w.createPullRequest(ctx, &taskMsg, result, issueNumber); err != nil {
			log.Printf("Failed to create PR for completed task: %v", err)

			// Handle PR creation errors
			taskErr := w.errorHandler.ClassifyError(err, &taskMsg)
			if taskErr.Recoverable {
				if handleErr := w.errorHandler.HandleTaskError(ctx, taskErr); handleErr != nil {
					log.Printf("PR error handling failed: %v", handleErr)
				}
			}
			// Still acknowledge the task as completed, PR creation is optional
		}
	}

	// Update task status to "completed" in database
	if updateErr := w.updateTaskStatusInDB(&taskMsg, "completed"); updateErr != nil {
		log.Printf("Warning: Failed to update task status to completed: %v", updateErr)
	}

	// Acknowledge successful processing
	msg.Ack(false)
	log.Printf("Task completed successfully for provider: %s", taskMsg.Provider)
}

// executeTaskConcurrent safely executes tasks with thread-safe failure tracking
func (w *ConcurrentTaskWorker) executeTaskConcurrent(ctx context.Context, taskMsg *TaskMessage) (*AITaskResponse, error) {
	log.Printf("Executing task with provider: %s (concurrent)", taskMsg.Provider)

	// Thread-safe failure count check and update
	w.mutex.Lock()

	// Check if enough time has passed to reset failure count
	if lastFailure, exists := w.lastFailureTime[taskMsg.Provider]; exists {
		if time.Since(lastFailure) > 10*time.Minute {
			log.Printf("Resetting failure count for provider %s due to time elapsed", taskMsg.Provider)
			w.failureCount[taskMsg.Provider] = 0
			delete(w.lastFailureTime, taskMsg.Provider)
		}
	}

	// Check failure count
	if w.failureCount[taskMsg.Provider] >= w.maxRetries {
		w.mutex.Unlock()
		log.Printf("Provider %s has failed %d times consecutively, skipping task",
			taskMsg.Provider, w.failureCount[taskMsg.Provider])
		w.recordTaskFailure(taskMsg, fmt.Sprintf("provider %s skipped due to too many consecutive failures", taskMsg.Provider))
		return nil, fmt.Errorf("provider %s skipped due to too many consecutive failures", taskMsg.Provider)
	}

	w.mutex.Unlock()

	// Get provider (this should be thread-safe)
	provider, exists := w.providerFactory.GetProvider(taskMsg.Provider)
	if !exists {
		w.mutex.Lock()
		w.failureCount[taskMsg.Provider]++
		w.mutex.Unlock()
		return nil, fmt.Errorf("unknown provider: %s", taskMsg.Provider)
	}

	// Check configuration
	if !provider.IsConfigured() {
		w.mutex.Lock()
		w.failureCount[taskMsg.Provider] += 2
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		w.mutex.Unlock()
		return nil, fmt.Errorf("provider %s is not properly configured", taskMsg.Provider)
	}

	// Convert and execute
	request := &AITaskRequest{
		Tasks:          taskMsg.Tasks,
		RepositoryName: taskMsg.RepositoryName,
		WorkingDir:     taskMsg.WorkingDir,
		Interactive:    taskMsg.Interactive,
		Cmd:            taskMsg.Cmd,
		ContinueTask:   taskMsg.ContinueTask,
		Timeout:        30 * time.Minute,
		Options:        make(map[string]interface{}),
	}

	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	response, err := provider.ExecuteTask(taskCtx, request)

	// Thread-safe failure count update
	w.mutex.Lock()
	if err != nil {
		// Update failure count based on error type
		if strings.Contains(err.Error(), "not properly configured") ||
		   strings.Contains(err.Error(), "API key") {
			w.failureCount[taskMsg.Provider] += 2
		} else {
			w.failureCount[taskMsg.Provider]++
		}
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		w.mutex.Unlock()
		return nil, fmt.Errorf("provider %s failed to execute task: %w", taskMsg.Provider, err)
	}

	if !response.Success {
		w.failureCount[taskMsg.Provider]++
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		w.mutex.Unlock()
		return response, fmt.Errorf("task execution failed: %s", response.Error)
	}

	// Reset failure count on success
	w.failureCount[taskMsg.Provider] = 0
	delete(w.lastFailureTime, taskMsg.Provider)
	w.mutex.Unlock()

	log.Printf("Provider %s task succeeded, failure count reset", taskMsg.Provider)
	return response, nil
}

// shouldCommitChanges determines if a completed task should commit changes
func (w *ConcurrentTaskWorker) shouldCommitChanges(result *AITaskResponse) bool {
	if result == nil || !result.Success {
		return false
	}

	// Check if task involved file changes
	return len(result.FilesModified) > 0 ||
		   strings.Contains(result.Output, "git add") ||
		   strings.Contains(result.Output, "modified:") ||
		   strings.Contains(result.Output, "new file:") ||
		   strings.Contains(result.Output, "Edit") ||
		   strings.Contains(result.Output, "Write") ||
		   strings.Contains(result.Output, "MultiEdit")
}

// shouldCreatePR determines if a completed task should create a PR
func (w *ConcurrentTaskWorker) shouldCreatePR(result *AITaskResponse) bool {
	if result == nil || !result.Success {
		return false
	}

	// Always create PR for successful tasks to ensure visibility
	// The PR creation process will validate if there are actual changes
	return true
}

// commitAndPushChanges commits and pushes changes to Git repository
func (w *ConcurrentTaskWorker) commitAndPushChanges(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse) error {
	log.Printf("Committing and pushing changes for task: %s", taskMsg.Tasks)

	// Use the working directory from taskMsg (already contains proper path like "JokerTrickster/gallery_ios")
	workingDir := fmt.Sprintf("../../git-repository/%s", taskMsg.WorkingDir)
	log.Printf("Using actual working directory for Git operations: %s", workingDir)

	// Validate that this is a Git repository (Claude provider should have set this up)
	gitDir := filepath.Join(workingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Printf("Working directory is not a Git repository: %s", workingDir)
		log.Printf("This indicates the Claude provider didn't set up the repository correctly")
		return nil // Don't fail the task, just skip Git operations
	}

	// Add all changes
	if err := w.runGitCommand(ctx, workingDir, "add", "."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Check if there are changes to commit
	statusOutput, err := w.runGitCommandWithOutput(ctx, workingDir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		log.Printf("No changes to commit for task: %s", taskMsg.Tasks)
		return nil
	}

	// Create commit message
	commitMsg := fmt.Sprintf("feat: %s\n\n🤖 Generated with [Claude Code](https://claude.ai/code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>", taskMsg.Tasks)

	// Commit changes
	if err := w.runGitCommand(ctx, workingDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Get current branch name
	branchOutput, err := w.runGitCommandWithOutput(ctx, workingDir, "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(branchOutput)

	// Push changes to the current branch and set upstream
	if err := w.runGitCommand(ctx, workingDir, "push", "--set-upstream", "origin", currentBranch); err != nil {
		return fmt.Errorf("failed to push changes to branch %s: %w", currentBranch, err)
	}

	log.Printf("Successfully committed and pushed changes for task: %s", taskMsg.Tasks)
	return nil
}

// runGitCommand executes a Git command in the specified directory
func (w *ConcurrentTaskWorker) runGitCommand(ctx context.Context, workingDir string, args ...string) error {
	_, err := w.runGitCommandWithOutput(ctx, workingDir, args...)
	return err
}

// runGitCommandWithOutput executes a Git command and returns output
func (w *ConcurrentTaskWorker) runGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Git command failed: git %s\nOutput: %s\nError: %v",
			strings.Join(args, " "), string(output), err)
		return string(output), err
	}

	log.Printf("Git command succeeded: git %s", strings.Join(args, " "))
	return string(output), nil
}

// cloneOrInitRepository clones a repository and creates a working branch
func (w *ConcurrentTaskWorker) cloneOrInitRepository(ctx context.Context, repositoryName, workingDir string) error {
	log.Printf("Setting up Git repository for %s in %s", repositoryName, workingDir)

	// Try to clone from GitHub first
	githubURL := fmt.Sprintf("https://github.com/%s.git", repositoryName)

	// Remove the existing directory and clone fresh
	if err := os.RemoveAll(workingDir); err != nil {
		log.Printf("Failed to remove existing directory: %v", err)
	}

	parentDir := filepath.Dir(workingDir)
	cloneName := filepath.Base(workingDir)

	// Try to clone the repository
	cmd := exec.CommandContext(ctx, "git", "clone", githubURL, cloneName)
	cmd.Dir = parentDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to clone repository %s: %v\nOutput: %s", githubURL, err, string(output))
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Printf("Successfully cloned repository %s to %s", githubURL, workingDir)

	// Create a new branch for this task
	branchName := fmt.Sprintf("task-%d", time.Now().Unix())
	if err := w.runGitCommand(ctx, workingDir, "checkout", "-b", branchName); err != nil {
		log.Printf("Warning: failed to create branch %s: %v", branchName, err)
		// Continue with main branch if branch creation fails
	} else {
		log.Printf("Created and switched to branch: %s", branchName)
	}

	return nil
}

// createPullRequest creates a GitHub PR for completed tasks
func (w *ConcurrentTaskWorker) createPullRequest(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse, issueNumber string) error {
	log.Printf("Creating PR for completed task: %s", taskMsg.Tasks)

	// Use absolute path for PR creation
	actualWorkingDir := GetRepositoryPath(taskMsg.RepositoryName)
	log.Printf("Starting PR creation process for task: %s", taskMsg.Tasks)
	log.Printf("Using actual working directory: %s", actualWorkingDir)

	// Create PR creator instance with actual repository directory
	prCreator := NewPRCreator(actualWorkingDir, taskMsg.RepositoryName)

	// Create the pull request with issue linking
	if err := prCreator.CreatePRForCompletedTaskWithIssue(ctx, taskMsg, result, issueNumber); err != nil {
		return fmt.Errorf("PR creation failed: %w", err)
	}

	log.Printf("Successfully created PR for task: %s", taskMsg.Tasks)
	return nil
}

// Close closes connections and waits for all tasks to complete
func (w *ConcurrentTaskWorker) Close() error {
	log.Println("Closing concurrent task worker...")

	// Wait for all running tasks to complete
	w.wg.Wait()

	if w.channel != nil {
		w.channel.Close()
	}
	if w.connection != nil {
		return w.connection.Close()
	}
	return nil
}

// Thread-safe getter methods
func (w *ConcurrentTaskWorker) GetAvailableProviders() []string {
	return w.providerFactory.GetAvailableProviders()
}

func (w *ConcurrentTaskWorker) GetFailureCountsReport() map[string]int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	report := make(map[string]int)
	for provider, count := range w.failureCount {
		report[provider] = count
	}
	return report
}

func (w *ConcurrentTaskWorker) ResetAllFailureCounts() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for provider := range w.failureCount {
		w.failureCount[provider] = 0
	}
	log.Println("Reset all provider failure counts (thread-safe)")
}

// createGitHubIssue creates a GitHub issue for the task
func (w *ConcurrentTaskWorker) createGitHubIssue(ctx context.Context, taskMsg *TaskMessage) (string, string, error) {
	if w.githubToken == "" || w.githubService == nil {
		log.Printf("GitHub integration not configured, skipping issue creation")
		return "", "", nil
	}

	// Parse repository name to extract owner and repo
	parts := strings.Split(taskMsg.RepositoryName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository name format: %s", taskMsg.RepositoryName)
	}
	owner, repo := parts[0], parts[1]

	// Generate request ID for tracking
	requestID := mysql.PKIDGenerate()

	// Create issue request
	issueBody := w.githubService.GenerateIssueTemplate(taskMsg.Tasks, taskMsg.RepositoryName, taskMsg.BranchName, requestID)
	issueReq := &response.CreateIssueRequest{
		Title: fmt.Sprintf("Task: %s", taskMsg.Tasks),
		Body:  &issueBody,
		Labels: []string{"automated-task", "claude-ai"},
	}

	// Create the issue
	issue, err := w.githubService.CreateIssue(ctx, w.githubToken, owner, repo, issueReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GitHub issue: %w", err)
	}

	log.Printf("Created GitHub issue #%d for task: %s", issue.Number, taskMsg.Tasks)
	return issue.HTMLURL, fmt.Sprintf("%d", issue.Number), nil
}

// recordTaskSuccess records successful task completion to database
func (w *ConcurrentTaskWorker) recordTaskSuccess(taskMsg *TaskMessage, result *AITaskResponse, issueURL string) {
	if w.db == nil {
		log.Printf("Database connection not available, skipping success record")
		return
	}

	requestID := mysql.PKIDGenerate()
	now := time.Now()
	processingTime := int64(0)

	// Serialize result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal task result: %v", err)
		resultJSON = []byte(`{"error": "failed to serialize result"}`)
	}
	resultStr := string(resultJSON)

	history := mysql.WorkflowHistories{
		RequestID:        requestID,
		Status:          "completed",
		Tasks:           taskMsg.Tasks,
		RepositoryName:  taskMsg.RepositoryName,
		Provider:        taskMsg.Provider,
		Interactive:     taskMsg.Interactive,
		ContinueTask:    taskMsg.ContinueTask,
		CreatedAt:       now,
		CompletedAt:     &now,
		ProcessingTimeMs: &processingTime,
		Result:          &resultStr,
	}

	if taskMsg.WorkingDir != "" {
		history.WorkingDir = &taskMsg.WorkingDir
	}
	if taskMsg.BranchName != "" {
		history.BranchName = &taskMsg.BranchName
	}
	if taskMsg.Cmd != "" {
		history.Cmd = &taskMsg.Cmd
	}
	if issueURL != "" {
		history.GitHubIssueURL = &issueURL
	}

	if err := w.db.Create(&history).Error; err != nil {
		log.Printf("Failed to record task success to database: %v", err)
	} else {
		log.Printf("Recorded task success to database with request_id: %s", requestID)
	}
}

// recordTaskFailure records task failure to database (thread-safe)
func (w *ConcurrentTaskWorker) recordTaskFailure(taskMsg *TaskMessage, errorMsg string) {
	if w.db == nil {
		log.Printf("Database connection not available, skipping failure record")
		return
	}

	requestID := mysql.PKIDGenerate()
	now := time.Now()
	processingTime := int64(0)

	history := mysql.WorkflowHistories{
		RequestID:        requestID,
		Status:          "failed",
		Tasks:           taskMsg.Tasks,
		RepositoryName:  taskMsg.RepositoryName,
		Provider:        taskMsg.Provider,
		Interactive:     taskMsg.Interactive,
		ContinueTask:    taskMsg.ContinueTask,
		CreatedAt:       now,
		CompletedAt:     &now,
		ProcessingTimeMs: &processingTime,
	}

	if taskMsg.WorkingDir != "" {
		history.WorkingDir = &taskMsg.WorkingDir
	}
	if taskMsg.BranchName != "" {
		history.BranchName = &taskMsg.BranchName
	}
	if taskMsg.Cmd != "" {
		history.Cmd = &taskMsg.Cmd
	}
	if errorMsg != "" {
		history.Error = &errorMsg
	}

	if err := w.db.Create(&history).Error; err != nil {
		log.Printf("Failed to record task failure to database: %v", err)
	} else {
		log.Printf("Recorded task failure to database with request_id: %s", requestID)
	}
}

// updateTaskStatusInDB updates the task status in database
func (w *ConcurrentTaskWorker) updateTaskStatusInDB(taskMsg *TaskMessage, status string) error {
	if w.db == nil {
		log.Printf("Database connection not available for status update")
		return nil // Don't fail task execution if DB is not available
	}

	// If no RequestID, try to find by tasks and repository name
	var requestID string
	if taskMsg.RequestID != "" {
		requestID = taskMsg.RequestID
	} else {
		// Fallback: find by tasks and repository name (not recommended for production)
		var workflow mysql.WorkflowHistories
		err := w.db.Where("tasks = ? AND repository_name = ? AND status IN (?)",
			taskMsg.Tasks, taskMsg.RepositoryName, []string{"pending", "processing"}).
			Order("created_at DESC").First(&workflow).Error
		if err != nil {
			log.Printf("Failed to find workflow by tasks and repository: %v", err)
			return err
		}
		requestID = workflow.RequestID
	}

	// Check if database is available
	if w.db == nil {
		log.Printf("Database not available, skipping status update")
		return nil
	}

	// Update status
	result := w.db.Model(&mysql.WorkflowHistories{}).
		Where("request_id = ?", requestID).
		Update("status", status)

	if result.Error != nil {
		log.Printf("Failed to update task status in database: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("No rows affected when updating status for request_id: %s", requestID)
		return fmt.Errorf("no task found with request_id: %s", requestID)
	}

	log.Printf("Updated task status to '%s' for request_id: %s", status, requestID)
	return nil
}