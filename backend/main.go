package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"local-backend-server/internal/infrastructure/config"
	"local-backend-server/internal/handlers"
	"local-backend-server/internal/infrastructure/database"
	"local-backend-server/internal/infrastructure/queue"
	"local-backend-server/internal/interfaces"
	"local-backend-server/internal/middleware"
	"local-backend-server/internal/services"
)

// Global variables for dependencies
var queuePublisher *queue.Publisher
var dbConnection *database.DB
var atomicService *services.AtomicQueueService
var taskHistoryHandler *handlers.TaskHistoryHandler

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		// Continue with default settings
		cfg = getDefaultConfig()
	}

	// Initialize database connection
	dbConn, err := database.NewConnection(cfg)
	if err != nil {
		log.Printf("Failed to create database connection: %v", err)
		log.Println("Continuing without database integration")
	} else {
		dbConnection = dbConn
		log.Println("Database connection initialized successfully")
	}

	// Initialize queue publisher
	if cfg.RabbitMQ.URL != "" {
		publisher, err := queue.NewPublisher(&cfg.RabbitMQ)
		if err != nil {
			log.Printf("Failed to create queue publisher: %v", err)
			log.Println("Continuing without queue integration")
		} else {
			queuePublisher = publisher
			log.Println("Queue publisher initialized successfully")
		}
	} else {
		log.Println("No RabbitMQ URL configured, continuing without queue integration")
	}

	// Initialize atomic service if both database and queue are available
	if dbConnection != nil {
		var publisher interfaces.MessagePublisher
		if queuePublisher != nil {
			publisher = queuePublisher
		}
		atomicService = services.NewAtomicQueueService(dbConnection.DB, publisher)
		log.Println("Atomic queue service initialized successfully")

		// Initialize task history handler
		taskHistoryHandler = handlers.NewTaskHistoryHandler(dbConnection.DB)
		log.Println("Task history handler initialized successfully")
	} else {
		log.Println("Database not available, atomic service and task history handler not initialized")
	}

	// Initialize Gin router
	r := gin.New()

	// Add comprehensive middleware stack
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.ErrorHandlingMiddleware())
	r.Use(middleware.ValidationErrorHandler())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.HealthCheckMiddleware())
	r.Use(middleware.TimeoutMiddleware(60 * time.Second))

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API routes group
	api := r.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.GET("/github", handleGitHubAuth)
			auth.GET("/github/callback", handleGitHubCallback)
			auth.POST("/logout", handleLogout)
		}

		// Repository routes
		repos := api.Group("/repos")
		repos.Use(authMiddleware())
		{
			repos.GET("/", handleGetRepos)
			repos.POST("/clone", handleCloneRepo)
			repos.GET("/:id/status", handleRepoStatus)
		}

		// Task routes
		tasks := api.Group("/tasks")
		tasks.Use(authMiddleware())
		{
			tasks.GET("/", handleGetTasks)
			tasks.POST("/", handleCreateTask)
			tasks.PUT("/:id", handleUpdateTask)
			tasks.DELETE("/:id", handleDeleteTask)
			tasks.POST("/:id/execute", handleExecuteTask)

			// Task history endpoint - only if handler is available
			if taskHistoryHandler != nil {
				tasks.GET("/history/:repository_name", taskHistoryHandler.GetTaskHistory)
			}
		}

		// AI routes
		ai := api.Group("/ai")
		ai.Use(authMiddleware())
		{
			ai.POST("/process", handleAIProcess)
			ai.GET("/tokens/status", handleTokenStatus)
		}

		// Claude routes
		claude := api.Group("/claude")
		claude.Use(authMiddleware())
		{
			claude.POST("/run-tasks", handleClaudeRunTasks)
		}

		// Notification routes
		notifications := api.Group("/notifications")
		notifications.Use(authMiddleware())
		{
			notifications.POST("/subscribe", handleSubscribeNotifications)
			notifications.POST("/send", handleSendNotification)
		}
		
		// Monitoring routes
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/health/detailed", handlers.HandleHealthDetailed)
			monitoring.GET("/metrics/errors", handlers.HandleErrorMetrics)
			monitoring.POST("/metrics/reset", handlers.HandleResetMetrics)
		}
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

// Placeholder handlers - will be implemented
func handleGitHubAuth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GitHub auth endpoint"})
}

func handleGitHubCallback(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GitHub callback endpoint"})
}

func handleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout endpoint"})
}

func handleGetRepos(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get repos endpoint"})
}

func handleCloneRepo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Clone repo endpoint"})
}

func handleRepoStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Repo status endpoint"})
}

func handleGetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get tasks endpoint"})
}

func handleCreateTask(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create task endpoint"})
}

func handleUpdateTask(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update task endpoint"})
}

func handleDeleteTask(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete task endpoint"})
}

func handleExecuteTask(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Execute task endpoint"})
}

func handleAIProcess(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "AI process endpoint"})
}

func handleTokenStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Token status endpoint"})
}

func handleSubscribeNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscribe notifications endpoint"})
}

func handleSendNotification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Send notification endpoint"})
}

// Claude task execution structures
type ReqRunTasksClaude struct {
	Tasks          string `json:"tasks" binding:"required"`           // 실행할 작업 내용
	RepositoryName string `json:"repository_name" binding:"required"` // 레포지토리 이름 (필수)
	WorkingDir     string `json:"working_dir,omitempty"`              // 작업 디렉토리 (옵션)
	Interactive    bool   `json:"interactive,omitempty"`              // 대화형 모드: 여러 작업을 순차 실행
	ClaudeCmd      string `json:"claude_cmd,omitempty"`               // Claude CLI 명령어 경로 (옵션)
	ContinueTask   bool   `json:"continue_task,omitempty"`            // 기존 작업 이어서 하기 (옵션)
}

type ClaudeTaskResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func handleClaudeRunTasks(c *gin.Context) {
	var req ReqRunTasksClaude
	if err := c.ShouldBindJSON(&req); err != nil {
		// Error handling is now done by middleware
		return
	}

	// Use atomic service if available, otherwise fall back to old behavior
	if atomicService != nil {
		// Generate session ID
		sessionID := fmt.Sprintf("session-%d", time.Now().Unix())

		// Create atomic publish request
		publishRequest := services.PublishRequest{
			Tasks:          req.Tasks,
			RepositoryName: req.RepositoryName,
			WorkingDir:     req.WorkingDir,
			Interactive:    req.Interactive,
			ClaudeCmd:      req.ClaudeCmd,
			ContinueTask:   req.ContinueTask,
			MessageType:    "claude_task",
			SessionID:      sessionID,
			Payload: map[string]interface{}{
				"request_type": "claude_task",
				"input": map[string]interface{}{
					"tasks":           req.Tasks,
					"repository_name": req.RepositoryName,
					"working_dir":     req.WorkingDir,
					"interactive":     req.Interactive,
					"claude_cmd":      req.ClaudeCmd,
					"continue_task":   req.ContinueTask,
				},
			},
		}

		// Execute atomic operation
		response, err := atomicService.PublishWithHistory(c.Request.Context(), publishRequest)
		if err != nil {
			// Use middleware error handling
			middleware.HandleError(c, err, "Failed to execute atomic Claude task operation")
			return
		}

		log.Printf("Claude task published atomically: %s - %s (ID: %s)", req.RepositoryName, req.Tasks, response.RequestID)

		// Convert response format
		claudeResponse := ClaudeTaskResponse{
			RequestID: response.RequestID,
			Status:    response.Status,
			Message:   response.Message,
			CreatedAt: response.CreatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusAccepted, claudeResponse)
		return
	}

	// Fallback to original implementation for backward compatibility
	log.Println("Atomic service not available, falling back to simple queue publishing")

	// Generate request ID (old timestamp format for compatibility)
	requestID := fmt.Sprintf("req-%d", time.Now().Unix())
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())

	// Publish to queue if available
	if queuePublisher != nil {
		// Create workflow message for queue
		workflowMessage := &queue.WorkflowMessage{
			Type:      "claude_task",
			ID:        requestID,
			SessionID: sessionID,
			Payload: map[string]interface{}{
				"request_type": "claude_task",
				"input": map[string]interface{}{
					"tasks":           req.Tasks,
					"repository_name": req.RepositoryName,
					"working_dir":     req.WorkingDir,
					"interactive":     req.Interactive,
					"claude_cmd":      req.ClaudeCmd,
					"continue_task":   req.ContinueTask,
				},
			},
			Timestamp: time.Now(),
		}

		// Publish message to queue
		if err := queuePublisher.PublishMessage(workflowMessage); err != nil {
			log.Printf("Failed to publish message to queue: %v", err)
			middleware.HandleError(c, err, "Failed to publish Claude task to queue (fallback mode)")
			return
		}

		log.Printf("Claude task published to queue: %s - %s", req.RepositoryName, req.Tasks)
	} else {
		log.Printf("Claude task logged (no queue): %s - %s", req.RepositoryName, req.Tasks)
	}

	response := ClaudeTaskResponse{
		RequestID: requestID,
		Status:    "pending",
		Message:   "Claude task has been queued for processing",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *config.Config {
	return &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			URL:            os.Getenv("RABBITMQ_URL"),
			QueueName:      os.Getenv("RABBITMQ_QUEUE_NAME"),
			MaxRetries:     3,
			RetryDelay:     time.Second * 5,
			ReconnectDelay: time.Second * 10,
			PrefetchCount:  1,
			Durable:        true,
			AutoDelete:     false,
		},
		Database: config.DatabaseConfig{
			Path:           "./data/workflow.db",
			MaxConnections: 10,
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
	}
}

// Auth middleware placeholder
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT token validation
		c.Next()
	}
}