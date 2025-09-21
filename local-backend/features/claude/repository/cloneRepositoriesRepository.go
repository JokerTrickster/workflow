package repository

import (
	_interface "main/features/claude/model/interface"

	"gorm.io/gorm"
)

func NewCloneRepositoriesRepository(gormDB *gorm.DB) _interface.ICloneRepositoriesRepository {
	return &CloneRepositoriesRepository{GormDB: gormDB}
}