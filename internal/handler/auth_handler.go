package handler

import (
	"errors"
	"net/http"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"

	"net/url"
	"os"

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

	utils.Success(c.Writer, response, "Registrasi berhasil. Silakan cek email untuk verifikasi.", http.StatusCreated)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	if token == "" {
		c.Redirect(http.StatusFound, frontendURL+"/sign-in?error=token_missing")
		return
	}

	if err := h.service.VerifyEmail(token); err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/sign-in?error="+url.QueryEscape(err.Error()))
		return
	}

	c.Redirect(http.StatusFound, frontendURL+"/sign-in?verified=true")
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

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.Success(c.Writer, nil, "Logout successful", http.StatusOK)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Token required", http.StatusBadRequest)
		return
	}

	response, err := h.service.GoogleLogin(req, c.Request.Context())
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Login berhasil", http.StatusOK)
}
