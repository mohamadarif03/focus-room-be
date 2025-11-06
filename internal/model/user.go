package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name  string `gorm:"type:varchar(100)" json:"name"`
	Email string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
}

type CreateUserInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
