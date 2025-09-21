package repository

import (
	"gorm.io/gorm"
)

type RunTasksClaudeRepository struct {
	GormDB *gorm.DB
}
