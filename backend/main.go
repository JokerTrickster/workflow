package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Gin router
	r := gin.Default()

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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "AI Git Workbench Backend is running",
		})
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
		}

		// AI routes
		ai := api.Group("/ai")
		ai.Use(authMiddleware())
		{
			ai.POST("/process", handleAIProcess)
			ai.GET("/tokens/status", handleTokenStatus)
		}

		// Notification routes
		notifications := api.Group("/notifications")
		notifications.Use(authMiddleware())
		{
			notifications.POST("/subscribe", handleSubscribeNotifications)
			notifications.POST("/send", handleSendNotification)
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

// Auth middleware placeholder
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT token validation
		c.Next()
	}
}