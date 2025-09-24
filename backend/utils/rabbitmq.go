package utils

import (
	"encoding/json"
	"fmt"
	"main/utils/config"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// TaskMessage matches the ReqRunTasksClaude structure from local-backend
type TaskMessage struct {
	Tasks          string `json:"tasks" validate:"required"`           // 실행할 작업 내용
	RepositoryName string `json:"repository_name" validate:"required"` // 레포지토리 이름 (필수)
	WorkingDir     string `json:"working_dir,omitempty"`               // 작업 디렉토리 (옵션)
	Interactive    bool   `json:"interactive,omitempty"`               // 대화형 모드: 여러 작업을 순차 실행
	Cmd            string `json:"cmd,omitempty"`                       // Claude CLI 명령어 경로 (옵션)
	ContinueTask   bool   `json:"continue_task,omitempty"`             // 기존 작업 이어서 하기 (옵션)
	Provider       string `json:"provider" validate:"required"`       // AI 모델 제공자 (예: "claude") (필수)
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
	conn, err := amqp.Dial(config.GlobalConfig.RabbitMQ.URL)
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
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",                                      // exchange
		config.GlobalConfig.RabbitMQ.QueueName, // routing key
		false,                                   // mandatory
		false,                                   // immediate
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