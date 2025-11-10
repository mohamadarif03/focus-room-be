package model

import (
	"time"
)

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Username       string    `gorm:"size:255;not null" json:"username"`
	Email          string    `gorm:"size:255;not null;unique" json:"email"`
	PasswordHash   string    `gorm:"size:255;not null" json:"-"`
	Role           string    `gorm:"size:50;not null" json:"role"`
	CurrentStreak  int       `gorm:"default:0" json:"current_streak"`
	KodePembimbing *string   `gorm:"size:50;unique;null" json:"kode_pembimbing"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Tasks []Task `gorm:"foreignKey:UserID" json:"tasks"`
}
