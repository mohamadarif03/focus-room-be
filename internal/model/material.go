package model

import (
	"time"
)

type Material struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `gorm:"not null"`
	Title         string `gorm:"size:255;not null"`
	SourceType    string `gorm:"size:50;not null"`
	Source        string `gorm:"size:255"`
	ExtractedText string `gorm:"type:text"`
	CreatedAt     time.Time

	User User `gorm:"foreignKey:UserID"`
}
