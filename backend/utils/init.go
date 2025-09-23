package utils

import (
	"fmt"
	"main/utils/config"
	"main/utils/db"
)

func InitServer() error {
	// 설정 초기화
	if err := config.InitConfig(); err != nil {
		fmt.Printf("config 초기화 에러 : %s", err.Error())
		return err
	}

	// 데이터베이스 초기화
	if err := db.InitDB(); err != nil {
		fmt.Printf("db 초기화 에러 : %s", err.Error())
		return err
	}

	// RabbitMQ 초기화 (선택적)
	if err := InitRabbitMQ(); err != nil {
		fmt.Printf("RabbitMQ 초기화 에러 : %s", err.Error())
		// RabbitMQ는 필수가 아니므로 에러를 반환하지 않고 계속 진행
	}

	return nil
}