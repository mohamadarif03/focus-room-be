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