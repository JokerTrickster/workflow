package utils

import (
	"fmt"
	_aws "main/utils/aws"
	"main/utils/db/mysql"
)

func InitServer() error {
	if err := _aws.InitAws(); err != nil {
		fmt.Printf("aws 초기화 에러 : %s\n", err.Error())
		// AWS 초기화 실패 시에도 계속 진행 (로컬 환경에서는 필요 없을 수 있음)
		fmt.Println("AWS 없이 계속 진행합니다...")
	}
	if err := mysql.InitMySQL(); err != nil {
		fmt.Printf("db 초기화 에러 : %s\n", err.Error())
		fmt.Println("데이터베이스 없이 계속 진행합니다... (일부 기능 제한됨)")
		// DB 초기화 실패 시에도 계속 진행 (태스크 실행은 가능)
	}
	return nil
}
