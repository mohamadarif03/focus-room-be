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

type AIHandler struct {
	service *service.AIService
}

func NewAIHandler(s *service.AIService) *AIHandler {
	return &AIHandler{service: s}
}

func (h *AIHandler) IngestPDF(c *gin.Context) {
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID", http.StatusInternalServerError)
		return
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		utils.Error(c.Writer, nil, "File 'pdf' tidak ditemukan", http.StatusBadRequest)
		return
	}

	title := c.PostForm("title")
	if title == "" {
		title = file.Filename
	}

	packageIDStr := c.PostForm("package_id")

	response, err := h.service.IngestPDF(c.Request.Context(), file, title, packageIDStr, userIDString.(string))
	if err != nil {
		if err.Error() == "package tidak ditemukan atau anda bukan pemiliknya" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "PDF ini tidak mengandung teks" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Materi PDF berhasil disimpan", http.StatusCreated)
}

func (h *AIHandler) IngestYouTube(c *gin.Context) {
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID", http.StatusInternalServerError)
		return
	}

	var req dto.IngestYouTubeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			formattedErrors := utils.FormatValidationError(validationErrs)
			utils.Error(c.Writer, formattedErrors, "Data tidak valid", http.StatusBadRequest)
		} else {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response, err := h.service.IngestYouTube(c.Request.Context(), req, userIDString.(string))
	if err != nil {
		if err.Error() == "package tidak ditemukan atau anda bukan pemiliknya" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "video ini tidak memiliki transkrip" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Materi YouTube berhasil disimpan", http.StatusCreated)
}

func (h *AIHandler) GenerateSummary(c *gin.Context) {
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID", http.StatusInternalServerError)
		return
	}

	var req dto.GenerateSummaryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			formattedErrors := utils.FormatValidationError(validationErrs)
			utils.Error(c.Writer, formattedErrors, "Data tidak valid", http.StatusBadRequest)
		} else {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response, err := h.service.GenerateSummary(c.Request.Context(), req, userIDString.(string))
	if err != nil {
		if err.Error() == "materi tidak ditemukan atau anda tidak punya akses" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Rangkuman berhasil dibuat", http.StatusOK)
}

func (h *AIHandler) GenerateQuiz(c *gin.Context) {
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID", http.StatusInternalServerError)
		return
	}

	var req dto.GenerateQuizRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			formattedErrors := utils.FormatValidationError(validationErrs)
			utils.Error(c.Writer, formattedErrors, "Data tidak valid", http.StatusBadRequest)
		} else {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		}
		return
	}
	response, err := h.service.GenerateQuiz(c.Request.Context(), req, userIDString.(string))
	if err != nil {
		if err.Error() == "materi tidak ditemukan atau anda tidak punya akses" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Kuis berhasil dibuat", http.StatusOK)
}
