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

func (r *MaterialRepository) Update(material *model.Material) error {
	return r.db.Save(material).Error
}

func (r *MaterialRepository) FindAll(userID uint, userRole string, packageID *uint) ([]model.Material, error) {
	var materials []model.Material

	query := r.db.Model(&model.Material{}).Preload("Package")

	if userRole == "admin" {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	}

	if packageID != nil {
		query = query.Where("package_id = ?", *packageID)
	}

	err := query.Order("created_at desc").Find(&materials).Error
	return materials, err
}

func (r *MaterialRepository) FindOneAccessible(id, userID uint, userRole string) (*model.Material, error) {
	var material model.Material

	query := r.db.Preload("Package").Where("id = ?", id)

	if userRole == "admin" {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	}

	err := query.First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
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

func (r *MaterialRepository) Delete(id, userID uint, userRole string) error {
	query := r.db.Where("id = ?", id)

	if userRole == "admin" {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	result := query.Unscoped().Delete(&model.Material{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
