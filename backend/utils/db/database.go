package db

import (
	"fmt"
	"main/utils/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	var err error

	// MySQL 데이터베이스 연결
	fmt.Printf("Connecting to MySQL at %s:%s/%s\n",
		config.GlobalConfig.Database.Host,
		config.GlobalConfig.Database.Port,
		config.GlobalConfig.Database.Name)

	DB, err = gorm.Open(mysql.Open(config.GlobalConfig.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 데이터베이스 마이그레이션
	if err := DB.AutoMigrate(&Task{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Println("Database connected and migrated successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}