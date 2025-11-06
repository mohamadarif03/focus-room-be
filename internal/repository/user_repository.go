package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.DB.Find(&users).Error
	return users, err
}
