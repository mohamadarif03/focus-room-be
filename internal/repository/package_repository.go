package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type PackageRepository struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) *PackageRepository {
	return &PackageRepository{db: db}
}

func (r *PackageRepository) Create(pkg *model.Package) (*model.Package, error) {
	err := r.db.Create(pkg).Error
	return pkg, err
}

func (r *PackageRepository) CreateWithMaterials(pkg *model.Package, materials []model.Material) error {
	tx := r.db.Begin()

	if err := tx.Create(pkg).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range materials {
		materials[i].PackageID = &pkg.ID
		materials[i].UserID = pkg.UserID

		if err := tx.Create(&materials[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *PackageRepository) FindByIDWithMaterials(id, userID uint) (*model.Package, error) {
	var pkg model.Package

	err := r.db.Preload("Materials", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc")
	}).Where("id = ?", id).First(&pkg).Error

	if pkg.IsPublic {
		return &pkg, err
	}

	if pkg.UserID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	return &pkg, err
}
func (r *PackageRepository) FindByID(id, userID uint) (*model.Package, error) {
	var pkg model.Package
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&pkg).Error
	return &pkg, err
}

func (r *PackageRepository) FindAllByUserID(userID uint) ([]model.Package, error) {
	var pkgs []model.Package

	err := r.db.Preload("Materials", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "package_id")
	}).Where("user_id = ?", userID).Order("created_at desc").Find(&pkgs).Error

	return pkgs, err
}

func (r *PackageRepository) Update(pkg *model.Package) (*model.Package, error) {
	err := r.db.Save(pkg).Error
	return pkg, err
}

func (r *PackageRepository) Delete(id, userID uint) error {
	err := r.db.Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&model.Package{}).Error
	return err
}
