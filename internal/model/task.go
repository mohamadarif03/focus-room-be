package model

import (
	"time"
)

type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Context   string    `gorm:"size:100;null" json:"context"`
	Priority  string    `gorm:"size:50;default:'medium'" json:"priority"`
	TaskDate  time.Time `gorm:"not null" json:"task_date"`
	Completed bool      `gorm:"default:false" json:"completed"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID"`
}
