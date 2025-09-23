package mysql

import (
	"time"

	"gorm.io/gorm"
)

// Workflow History Model (단순화된 버전)
// 프론트엔드에서 보여주는 정보만 저장
type WorkflowHistories struct {
	ID        uint64 `json:"id" gorm:"primaryKey;column:id"`
	RequestID string `json:"request_id" gorm:"column:request_id;uniqueIndex"`
	Status    string `json:"status" gorm:"column:status;index;default:pending"`

	// 요청 정보 (프론트엔드 입력값)
	Tasks          string  `json:"tasks" gorm:"column:tasks;type:text;not null"`
	RepositoryName string  `json:"repository_name" gorm:"column:repository_name;index;not null"`
	WorkingDir     *string `json:"working_dir,omitempty" gorm:"column:working_dir"`
	Cmd            *string `json:"cmd,omitempty" gorm:"column:cmd"`
	Interactive    bool    `json:"interactive" gorm:"column:interactive;default:false"`
	ContinueTask   bool    `json:"continue_task" gorm:"column:continue_task;default:false"`
	Provider       string  `json:"provider" gorm:"column:provider;not null"`

	// 실행 시간 정보
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at;index"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	ProcessingTimeMs *int64     `json:"processing_time_ms,omitempty" gorm:"column:processing_time_ms"`

	// 결과 정보
	Result *string `json:"result,omitempty" gorm:"column:result;type:text"`
	Error  *string `json:"error,omitempty" gorm:"column:error;type:text"`
}

func (WorkflowHistories) TableName() string {
	return "workflow_histories"
}

// 기존 구조체들
type Times struct {
	gorm.Model
	Timer       uint   `json:"timer" gorm:"column:timer"`
	Description string `json:"description" gorm:"column:description"`
}

type Tokens struct {
	gorm.Model
	UserID           uint   `json:"userID" gorm:"column:user_id"`
	AccessToken      string `json:"accessToken" gorm:"column:access_token"`
	RefreshToken     string `json:"refreshToken" gorm:"column:refresh_token"`
	RefreshExpiredAt int64  `json:"refreshExpiredAt" gorm:"column:refresh_expired_at"`
}

type Users struct {
	gorm.Model
	Name         string `json:"name" gorm:"column:name"`
	Email        string `json:"email" gorm:"uniqueIndex;column:email"`
	Password     string `json:"password" gorm:"column:password"`
	Coin         int    `json:"coin" gorm:"column:coin"`
	State        string `json:"state" gorm:"column:state"` //logout, wait, play
	RoomID       int    `json:"roomID" gorm:"column:room_id"`
	Provider     string `json:"provider" gorm:"column:provider"`
	ProfileID    int    `json:"profileID" gorm:"column:profile_id"`
	AlertEnabled bool   `json:"alertEnabled" gorm:"column:alert_enabled"`
}
