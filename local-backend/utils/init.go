package utils

import (
	"fmt"
	_aws "main/utils/aws"
	"main/utils/db/mysql"
)

func InitServer() error {
	if err := _aws.InitAws(); err != nil {
		fmt.Sprintf("aws 초기화 에러 : %s", err.Error())
		return err
	}
	if err := mysql.InitMySQL(); err != nil {
		fmt.Sprintf("db 초기화 에러 : %s", err.Error())
		return err
	}
	return nil
}
