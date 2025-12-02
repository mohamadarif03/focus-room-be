package model

import (
	"time"
)

type MaterialChat struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MaterialID uint      `gorm:"not null" json:"material_id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Role       string    `gorm:"size:20;not null" json:"role"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	CreatedAt  time.Time `json:"created_at"`

	Material Material `gorm:"foreignKey:MaterialID"`
	User     User     `gorm:"foreignKey:UserID"`
}
