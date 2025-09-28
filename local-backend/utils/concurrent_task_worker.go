package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Execute task
	result, err := w.executeTaskConcurrent(ctx, &taskMsg)
	if err != nil {
		log.Printf("Failed to execute task: %v", err)
		w.recordTaskFailure(&taskMsg, err.Error())
		msg.Ack(false)
		return
	}

	// Handle successful completion - commit and push changes first
	if w.shouldCommitChanges(result) {
		if err := w.commitAndPushChanges(ctx, &taskMsg, result); err != nil {
			log.Printf("Failed to commit and push changes: %v", err)
			// Continue with PR creation even if commit/push fails
		}
	}

	// Check if PR needed after committing changes
	if w.shouldCreatePR(result) {
		if err := w.createPullRequest(ctx, &taskMsg, result); err != nil {
			log.Printf("Failed to create PR for completed task: %v", err)
			// Still acknowledge the task as completed, PR creation is optional
		}
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

	workingDir := taskMsg.WorkingDir
	if workingDir == "" {
		return fmt.Errorf("working directory is required for Git operations")
	}

	// Check if it's a Git repository, if not try to initialize or clone
	gitDir := filepath.Join(workingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Printf("Not a Git repository: %s", workingDir)

		// Try to clone the repository if repository name is provided
		if taskMsg.RepositoryName != "" {
			if err := w.cloneOrInitRepository(ctx, taskMsg.RepositoryName, workingDir); err != nil {
				log.Printf("Failed to setup Git repository: %v", err)
				return nil // Don't fail the task, just skip Git operations
			}
		} else {
			log.Printf("No repository name provided, skipping Git operations")
			return nil
		}
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
func (w *ConcurrentTaskWorker) createPullRequest(ctx context.Context, taskMsg *TaskMessage, result *AITaskResponse) error {
	log.Printf("Creating PR for completed task: %s", taskMsg.Tasks)

	// Create PR creator instance
	prCreator := NewPRCreator(taskMsg.WorkingDir, taskMsg.RepositoryName)

	// Create the pull request
	if err := prCreator.CreatePRForCompletedTask(ctx, taskMsg, result); err != nil {
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