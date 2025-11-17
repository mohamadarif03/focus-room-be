package model

import "time"

type Package struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"size:255;not null"`
	ColorIcon string `gorm:"size:50;not null"`
	UserID    uint   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User      User       `gorm:"foreignKey:UserID"`
	Materials []Material `gorm:"foreignKey:PackageID"`
}
