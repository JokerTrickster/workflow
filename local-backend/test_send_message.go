package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"main/utils"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load RabbitMQ configuration
	config := utils.LoadRabbitMQConfigFromEnv()
	log.Printf("Connecting to RabbitMQ: %s", config.URL)

	// Create connection
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare queue
	q, err := ch.QueueDeclare(
		config.QueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Create test message
	message := utils.WorkflowMessage{
		Type:      "claude_task",
		ID:        uuid.New().String(),
		SessionID: "test_session_" + time.Now().Format("20060102_150405"),
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"input": map[string]interface{}{
				"repository_name": "gallery_ios",
				"tasks":          "앱 메인 화면을 만들거다. 메인 페이지에는 업로드 화면과 다운로드할 수 있는 화면이 보여야 된다. 추가해줘",
				"working_dir":    "",
				"interactive":    false,
				"continue_task":  false,
			},
		},
	}

	// Convert to JSON
	body, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	// Publish message
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}

	log.Printf("✅ Message sent successfully!")
	log.Printf("📨 Message ID: %s", message.ID)
	log.Printf("📦 Queue: %s", q.Name)
	log.Printf("🎯 Repository: %s", "gallery_ios")
	log.Printf("📝 Tasks: %s", "앱 메인 화면을 만들거다...")
	log.Printf("🕐 Timestamp: %s", message.Timestamp.Format("2006-01-02 15:04:05"))
}