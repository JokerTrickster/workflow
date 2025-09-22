package main

import (
	"log"

	"github.com/joho/godotenv"

	"main/utils"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load RabbitMQ configuration from environment
	config := utils.LoadRabbitMQConfigFromEnv()

	log.Printf("Testing RabbitMQ connection to: %s", config.Host)
	log.Printf("Queue name: %s", config.QueueName)

	// Test connection
	if err := utils.TestRabbitMQConnection(config); err != nil {
		log.Fatalf("RabbitMQ connection test failed: %v", err)
	}

	log.Println("✅ RabbitMQ connection test passed!")

	// Create connection for consuming messages
	conn, err := utils.NewRabbitMQConnection(config)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ connection: %v", err)
	}
	defer conn.Close()

	log.Println("🎉 RabbitMQ setup complete! Ready to consume messages.")
	log.Println("Press Ctrl+C to exit...")

	// Test message handler
	messageHandler := func(msg utils.WorkflowMessage) error {
		log.Printf("📨 Received message:")
		log.Printf("  Type: %s", msg.Type)
		log.Printf("  ID: %s", msg.ID)
		log.Printf("  SessionID: %s", msg.SessionID)
		log.Printf("  Timestamp: %s", msg.Timestamp.Format("2006-01-02 15:04:05"))

		// Extract task details from payload
		if msg.Type == "claude_task" {
			if payload, ok := msg.Payload["input"].(map[string]interface{}); ok {
				if tasks, taskOk := payload["tasks"].(string); taskOk {
					log.Printf("  Tasks: %s", tasks)
				}
				if repo, repoOk := payload["repository_name"].(string); repoOk {
					log.Printf("  Repository: %s", repo)
				}
			}
		}

		log.Println("✅ Message processed successfully!")
		return nil
	}

	// Start consuming messages
	log.Println("🔄 Starting message consumption...")
	if err := conn.ConsumeMessages(messageHandler); err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}
}