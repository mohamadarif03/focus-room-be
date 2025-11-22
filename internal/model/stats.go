package model

import (
	"time"
)

type FocusSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Duration  int       `gorm:"not null" json:"duration"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID"`
}

type QuizLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Score     int       `gorm:"not null" json:"score"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID"`
}
