package handler

import (
	"errors"
	"net/http"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

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

	response, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "email already registered" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusConflict)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "User registered successfully", http.StatusCreated)
}


func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

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

	response, err := h.service.Login(req)
	if err != nil {
		if err.Error() == "email atau password salah" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusUnauthorized)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Login berhasil", http.StatusOK)
}
