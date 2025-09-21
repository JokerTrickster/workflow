package repository

import (
	_interface "main/features/claude/model/interface"

	"gorm.io/gorm"
)

func NewRunTasksClaudeRepository(gormDB *gorm.DB) _interface.IRunTasksClaudeRepository {
	return &RunTasksClaudeRepository{GormDB: gormDB}
}
