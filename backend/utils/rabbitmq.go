package utils

import (
	"encoding/json"
	"fmt"
	"main/utils/config"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

type TaskMessage struct {
	RequestID      string                 `json:"request_id"`
	Type           string                 `json:"type"`
	Tasks          string                 `json:"tasks"`
	RepositoryName string                 `json:"repository_name"`
	WorkingDir     *string                `json:"working_dir,omitempty"`
	Cmd            *string                `json:"cmd,omitempty"`
	Provider       string                 `json:"provider"`
	Interactive    bool                   `json:"interactive"`
	Payload        map[string]interface{} `json:"payload"`
	Timestamp      time.Time              `json:"timestamp"`
}

var GlobalRabbitMQ *RabbitMQClient

func InitRabbitMQ() error {
	var err error
	GlobalRabbitMQ, err = NewRabbitMQClient()
	if err != nil {
		fmt.Printf("RabbitMQ 연결 실패: %v, 큐 없이 계속 진행\n", err)
		return nil // RabbitMQ는 선택적이므로 에러를 반환하지 않음
	}
	return nil
}

func NewRabbitMQClient() (*RabbitMQClient, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.GlobalConfig.RabbitMQ.Username, config.GlobalConfig.RabbitMQ.Password, config.GlobalConfig.RabbitMQ.Host, config.GlobalConfig.RabbitMQ.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 큐 선언
	_, err = ch.QueueDeclare(
		config.GlobalConfig.RabbitMQ.QueueName, // queue name
		true,                                   // durable
		false,                                  // delete when unused
		false,                                  // exclusive
		false,                                  // no-wait
		nil,                                    // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQClient{
		connection: conn,
		channel:    ch,
	}, nil
}

func (r *RabbitMQClient) PublishTaskMessage(msg *TaskMessage) error {
	if r == nil || r.channel == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	// 타임스탬프 설정
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",                                     // exchange
		config.GlobalConfig.RabbitMQ.QueueName, // routing key
		false,                                  // mandatory
		false,                                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		return r.connection.Close()
	}
	return nil
}

func PublishTask(msg *TaskMessage) error {
	if GlobalRabbitMQ == nil {
		return fmt.Errorf("RabbitMQ not available")
	}
	return GlobalRabbitMQ.PublishTaskMessage(msg)
}
