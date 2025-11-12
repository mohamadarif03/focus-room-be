package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.Success(c.Writer, users, "Berhasil mengambil data users", http.StatusOK)
}

func (h *UserHandler) GetSelf(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.service.GetSelf(userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, user, "Berhasil mengambil data profil", http.StatusOK)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c.Writer, nil, "ID user tidak valid", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, user, "Berhasil mengambil data user", http.StatusOK)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c.Writer, nil, "ID user tidak valid", http.StatusBadRequest)
		return
	}

	var req dto.UpdateUserRequest
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

	user, err := h.service.UpdateUser(uint(id), req)
	if err != nil {
		if err.Error() == "email sudah terdaftar" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusConflict)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, user, "Berhasil memperbarui user", http.StatusOK)
}

func (h *UserHandler) CheckAndUpdateStreak(c *gin.Context) {
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID dari token", http.StatusInternalServerError)
		return
	}

	_, err := h.service.CheckAndUpdateStreak(userIDString.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := h.service.GetSelf(userIDString.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}

	utils.Success(c.Writer, response, "Streak berhasil dicek dan diupdate", http.StatusOK)
}
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c.Writer, nil, "ID user tidak valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteUser(uint(id))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}
	utils.Success(c.Writer, nil, "Berhasil menghapus user", http.StatusOK)
}
