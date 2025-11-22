package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type PackageHandler struct {
	service *service.PackageService
}

func NewPackageHandler(s *service.PackageService) *PackageHandler {
	return &PackageHandler{service: s}
}

func (h *PackageHandler) CreatePackage(c *gin.Context) {
	userIDString, _ := c.Get("user_id")
	var req dto.PackageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			formattedErrors := utils.FormatValidationError(validationErrs)
			utils.Error(c.Writer, formattedErrors, "Data yang diberikan tidak valid", http.StatusBadRequest)
		} else {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response, err := h.service.CreatePackage(req, userIDString.(string))

	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Package berhasil dibuat", http.StatusCreated)
}
func (h *PackageHandler) GetMyPackages(c *gin.Context) {
	userIDString, _ := c.Get("user_id")
	response, err := h.service.GetMyPackages(userIDString.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.Success(c.Writer, response, "Berhasil mengambil packages", http.StatusOK)
}

func (h *PackageHandler) GetPackageByID(c *gin.Context) {
	idStr := c.Param("id")
	userID, _ := c.Get("user_id")

	response, err := h.service.GetPackageWithMaterials(idStr, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}

	utils.Success(c.Writer, response, "Berhasil mengambil detail package", http.StatusOK)
}

func (h *PackageHandler) UpdatePackage(c *gin.Context) {
	userIDString, _ := c.Get("user_id")
	idStr := c.Param("id")
	var req dto.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			formattedErrors := utils.FormatValidationError(validationErrs)
			utils.Error(c.Writer, formattedErrors, "Data yang diberikan tidak valid", http.StatusBadRequest)
		} else {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response, err := h.service.UpdatePackage(idStr, userIDString.(string), req)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, response, "Package berhasil diupdate", http.StatusOK)
}

func (h *PackageHandler) AdminCreatePackage(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.AdminCreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.service.AdminCreatePackage(req, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Package kurikulum berhasil dibuat", http.StatusCreated)
}

func (h *PackageHandler) DeletePackage(c *gin.Context) {
	userIDString, _ := c.Get("user_id")
	idStr := c.Param("id")
	err := h.service.DeletePackage(idStr, userIDString.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, nil, "Package berhasil dihapus", http.StatusOK)
}
