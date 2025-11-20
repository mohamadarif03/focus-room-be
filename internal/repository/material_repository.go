package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type MaterialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

func (r *MaterialRepository) Save(material *model.Material) (*model.Material, error) {
	err := r.db.Create(&material).Error
	return material, err
}

func (r *MaterialRepository) FindByID(id, userID uint) (*model.Material, error) {
	var material model.Material
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&material).Error
	return &material, err
}

func (r *MaterialRepository) FindRandomByUserID(userID uint, limit int) ([]model.Material, error) {
	var materials []model.Material
	err := r.db.Where("user_id = ?", userID).Order("RANDOM()").Limit(limit).Find(&materials).Error
	return materials, err
}

func (r *MaterialRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Material{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
