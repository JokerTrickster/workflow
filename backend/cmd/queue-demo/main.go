package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/internal/infrastructure/queue"
)

func main() {
	log.Println("🚀 RabbitMQ Queue Integration Demo")

	// Load configuration
	cfg := config.Load()
	
	// Create RabbitMQ repository
	queueRepo, err := queue.NewRabbitMQRepository(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("❌ Failed to create RabbitMQ repository: %v", err)
	}
	defer func() {
		if closer, ok := queueRepo.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	ctx := context.Background()

	// Demonstrate health check
	log.Println("\n📊 Checking queue health...")
	health, err := queueRepo.GetQueueHealth(ctx)
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
	} else {
		log.Printf("🏥 Queue Health: Healthy=%v, Issues=%v", health.IsHealthy, health.Issues)
	}

	// Demonstrate statistics
	log.Println("\n📈 Getting queue statistics...")
	stats, err := queueRepo.GetQueueStatistics(ctx)
	if err != nil {
		log.Printf("❌ Failed to get statistics: %v", err)
	} else {
		log.Printf("📊 Statistics: Enqueued=%d, Failed=%d, QueueLength=%d", 
			stats.TotalEnqueued, stats.TotalFailed, stats.CurrentQueueLength)
	}

	// Create a sample task
	log.Println("\n📦 Creating sample task...")
	taskID := valueobjects.GenerateTaskID()
	userID, _ := valueobjects.NewUserID("demo-user")
	repoPath, _ := valueobjects.NewRepositoryPath("example/repository")
	branchName, _ := valueobjects.NewBranchName("feature/queue-integration")

	task, err := entities.NewTask(
		taskID,
		userID,
		"RabbitMQ Integration Demo Task",
		"This is a demonstration task to test RabbitMQ integration with the task queue system.",
		repoPath,
		"queue-integration-epic",
		branchName,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create task: %v", err)
	}

	log.Printf("✅ Created task: %s", task.ID().Value())

	// Demonstrate task enqueueing
	log.Println("\n📨 Enqueueing task to RabbitMQ...")
	err = queueRepo.Enqueue(ctx, task)
	if err != nil {
		log.Printf("❌ Failed to enqueue task: %v", err)
	} else {
		log.Printf("✅ Successfully enqueued task: %s", task.ID().Value())
	}

	// Wait a moment and check statistics again
	time.Sleep(1 * time.Second)
	log.Println("\n📈 Getting updated statistics...")
	stats, err = queueRepo.GetQueueStatistics(ctx)
	if err != nil {
		log.Printf("❌ Failed to get updated statistics: %v", err)
	} else {
		log.Printf("📊 Updated Statistics: Enqueued=%d, Failed=%d, QueueLength=%d", 
			stats.TotalEnqueued, stats.TotalFailed, stats.CurrentQueueLength)
	}

	// Try to get queue length directly
	log.Println("\n📏 Checking queue length...")
	length, err := queueRepo.GetQueueLength(ctx)
	if err != nil {
		log.Printf("❌ Failed to get queue length: %v", err)
	} else {
		log.Printf("📏 Current queue length: %d", length)
	}

	// Demonstrate batch enqueueing
	log.Println("\n📦 Creating batch of tasks...")
	var batchTasks []*entities.Task
	for i := 0; i < 3; i++ {
		taskID := valueobjects.GenerateTaskID()
		batchTask, err := entities.NewTask(
			taskID,
			userID,
			fmt.Sprintf("Batch Task %d", i+1),
			fmt.Sprintf("This is batch task number %d", i+1),
			repoPath,
			"queue-integration-epic",
			branchName,
		)
		if err != nil {
			log.Printf("❌ Failed to create batch task %d: %v", i+1, err)
			continue
		}
		batchTasks = append(batchTasks, batchTask)
	}

	log.Printf("📦 Created %d batch tasks", len(batchTasks))

	// Enqueue batch
	log.Println("\n📨 Enqueueing batch tasks...")
	err = queueRepo.EnqueueBatch(ctx, batchTasks)
	if err != nil {
		log.Printf("❌ Batch enqueue had errors: %v", err)
	} else {
		log.Printf("✅ Successfully enqueued %d batch tasks", len(batchTasks))
	}

	// Final statistics
	time.Sleep(1 * time.Second)
	log.Println("\n📈 Final statistics...")
	stats, err = queueRepo.GetQueueStatistics(ctx)
	if err != nil {
		log.Printf("❌ Failed to get final statistics: %v", err)
	} else {
		log.Printf("📊 Final Statistics: Enqueued=%d, Failed=%d, QueueLength=%d", 
			stats.TotalEnqueued, stats.TotalFailed, stats.CurrentQueueLength)
	}

	// Try queue length one more time
	length, err = queueRepo.GetQueueLength(ctx)
	if err != nil {
		log.Printf("❌ Failed to get final queue length: %v", err)
	} else {
		log.Printf("📏 Final queue length: %d", length)
	}

	log.Println("\n🎉 Demo completed!")
	
	// Show usage instructions
	log.Println("\n💡 Usage Notes:")
	log.Println("   - Set USE_RABBITMQ=true to enable RabbitMQ in the application")
	log.Println("   - Configure RabbitMQ connection via environment variables:")
	log.Println("     RABBITMQ_HOST=13.203.37.93")
	log.Println("     RABBITMQ_PORT=5672")
	log.Println("     RABBITMQ_QUEUE=claude-tasks")
	log.Println("   - Check RabbitMQ Management UI at http://13.203.37.93:15672/")
	
	if os.Getenv("RABBITMQ_HOST") == "" {
		log.Println("\n⚠️  Note: Using default RabbitMQ config. Set environment variables for production use.")
	}
}