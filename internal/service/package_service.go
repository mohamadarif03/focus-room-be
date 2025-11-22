package service

import (
	"errors"
	"strconv"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
)

type PackageService struct {
	pkgRepo *repository.PackageRepository
}

func NewPackageService(pkgRepo *repository.PackageRepository) *PackageService {
	return &PackageService{pkgRepo: pkgRepo}
}

func packageToResponse(pkg *model.Package) dto.PackageResponse {
	return dto.PackageResponse{
		ID:        pkg.ID,
		Title:     pkg.Title,
		ColorIcon: pkg.ColorIcon,
		UserID:    pkg.UserID,
		CreatedAt: pkg.CreatedAt,
	}
}

func (s *PackageService) CreatePackage(req dto.PackageRequest, userIDString string) (*dto.PackageResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	newPkg := &model.Package{
		Title:     req.Title,
		ColorIcon: req.ColorIcon,
		UserID:    uint(userID),
	}

	createdPkg, err := s.pkgRepo.Create(newPkg)
	if err != nil {
		return nil, errors.New("gagal membuat package")
	}

	response := packageToResponse(createdPkg)
	return &response, nil
}

func (s *PackageService) AdminCreatePackage(req dto.AdminCreatePackageRequest, adminIDString string) (*dto.PackageWithMaterialsResponse, error) {
	adminID, _ := strconv.ParseUint(adminIDString, 10, 32)

	newPkg := model.Package{
		Title:     req.Title,
		ColorIcon: req.ColorIcon,
		UserID:    uint(adminID),
		IsPublic:  req.IsPublic,
	}

	var materials []model.Material
	var materialsResponse []dto.MaterialSimple

	for _, mReq := range req.Materials {
		mat := model.Material{
			Title:      mReq.Title,
			SourceType: mReq.SourceType,
			Source:     mReq.Source,
		}
		materials = append(materials, mat)

		materialsResponse = append(materialsResponse, dto.MaterialSimple{
			Title:      mReq.Title,
			SourceType: mReq.SourceType,
		})
	}

	if err := s.pkgRepo.CreateWithMaterials(&newPkg, materials); err != nil {
		return nil, errors.New("gagal membuat package dan materi")
	}

	return &dto.PackageWithMaterialsResponse{
		ID:        newPkg.ID,
		Title:     newPkg.Title,
		ColorIcon: newPkg.ColorIcon,
		Materials: materialsResponse,
	}, nil
}


func (s *PackageService) GetPackageWithMaterials(idString, userIDString string) (*dto.PackageWithMaterialsResponse, error) {
	id, err := strconv.ParseUint(idString, 10, 32)
	if err != nil {
		return nil, errors.New("ID package tidak valid")
	}
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	pkg, err := s.pkgRepo.FindByIDWithMaterials(uint(id), uint(userID))
	if err != nil {
		return nil, errors.New("package tidak ditemukan atau anda tidak memiliki akses")
	}

	var materialResponses []dto.MaterialSimple
	for _, m := range pkg.Materials {
		materialResponses = append(materialResponses, dto.MaterialSimple{
			ID:         m.ID,
			Title:      m.Title,
			SourceType: m.SourceType,
		})
	}

	return &dto.PackageWithMaterialsResponse{
		ID:        pkg.ID,
		Title:     pkg.Title,
		ColorIcon: pkg.ColorIcon,
		Materials: materialResponses,
	}, nil
}
func (s *PackageService) GetMyPackages(userIDString string) ([]dto.PackageResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	pkgs, err := s.pkgRepo.FindAllByUserID(uint(userID))
	if err != nil {
		return nil, errors.New("gagal mengambil packages")
	}

	var responses []dto.PackageResponse
	for _, pkg := range pkgs {
		responses = append(responses, packageToResponse(&pkg))
	}
	return responses, nil
}


func (s *PackageService) UpdatePackage(idStr, userIDString string, req dto.PackageRequest) (*dto.PackageResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)
	pkgID, _ := strconv.ParseUint(idStr, 10, 32)

	pkg, err := s.pkgRepo.FindByID(uint(pkgID), uint(userID))
	if err != nil {
		return nil, errors.New("package tidak ditemukan")
	}

	pkg.Title = req.Title
	pkg.ColorIcon = req.ColorIcon

	updatedPkg, err := s.pkgRepo.Update(pkg)
	if err != nil {
		return nil, errors.New("gagal update package")
	}

	response := packageToResponse(updatedPkg)
	return &response, nil
}

func (s *PackageService) DeletePackage(idStr, userIDString string) error {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)
	pkgID, _ := strconv.ParseUint(idStr, 10, 32)

	_, err := s.pkgRepo.FindByID(uint(pkgID), uint(userID))
	if err != nil {
		return errors.New("package tidak ditemukan")
	}

	return s.pkgRepo.Delete(uint(pkgID), uint(userID))
}
