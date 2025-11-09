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

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.DB.Find(&users).Error
	return users, err
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("email = ?", email).Where("role != ?", "admin").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB.Where("ID = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) (*model.User, error) {
	err := r.DB.Save(&user).Error
	return user, err
}

func (r *UserRepository) Delete(id uint) error {
	err := r.DB.Unscoped().Where("ID = ?", id).Delete(&model.User{}).Error
	return err
}

func (r *UserRepository) CreateUser(user *model.User) (*model.User, error) {
	err := r.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
