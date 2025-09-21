package mysql

import (
	"gorm.io/gorm"
)

// 전체, 한식, 중식, 일식, 양식, 분식, 패스트푸드, 카페, 술집, 기타
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
