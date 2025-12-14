package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

// POST /materials/pdf
func (h *AIHandler) IngestPDF(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.Error(c.Writer, nil, "File PDF wajib diupload", http.StatusBadRequest)
		return
	}

	pkgID := c.PostForm("package_id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role") // Ambil Role untuk penanda IsPublic

	resp, err := h.service.IngestPDF(c.Request.Context(), fileHeader, pkgID, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil memproses PDF", http.StatusCreated)
}

func (h *AIHandler) GetMaterialDetail(c *gin.Context) {
	idStr := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	resp, err := h.service.GetMaterialDetail(idStr, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil detail materi", http.StatusOK)
}

func (h *AIHandler) IngestYouTube(c *gin.Context) {
	var req dto.IngestYouTubeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	resp, err := h.service.IngestYouTube(c.Request.Context(), req, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil memproses YouTube", http.StatusCreated)
}

// GET /materials?package_id=1
func (h *AIHandler) GetMaterials(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	packageID := c.Query("package_id")

	resp, err := h.service.GetMaterials(userID.(string), role.(string), packageID)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil daftar materi", http.StatusOK)
}

// POST /ai/summarize
func (h *AIHandler) GenerateSummary(c *gin.Context) {
	var req dto.GenerateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.service.GenerateSummary(c.Request.Context(), req, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Rangkuman berhasil dibuat", http.StatusOK)
}

func (h *AIHandler) GenerateQuiz(c *gin.Context) {
	var req dto.GenerateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.service.GenerateQuiz(c.Request.Context(), req, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Quiz berhasil dibuat", http.StatusOK)
}

// POST /ai/flashcards
func (h *AIHandler) GenerateFlashcards(c *gin.Context) {
	var req dto.GenerateFlashcardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Format data salah atau count tidak valid (min 1, max 20)", http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.service.GenerateFlashcards(c.Request.Context(), req, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Flashcards berhasil dibuat", http.StatusOK)
}

// POST /daily-quiz
func (h *AIHandler) GenerateDailyQuiz(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.GenerateDailyQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Missing mode or topic", http.StatusBadRequest)
		return
	}

	resp, err := h.service.GenerateDailyQuiz(c.Request.Context(), userID.(string), req)
	if err != nil {
		if strings.Contains(err.Error(), "anda belum memiliki materi") {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil Daily Quiz", http.StatusOK)
}

func (h *AIHandler) ClaimDailyStreak(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.ClaimQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	if req.Score != 10 {
		utils.Error(c.Writer, nil, "Anda harus menjawab semua soal dengan benar untuk klaim streak!", http.StatusBadRequest)
		return
	}

	err := h.service.ClaimDailyStreak(userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusConflict)
		return
	}

	utils.Success(c.Writer, nil, "Selamat! Jawaban sempurna. Streak bertambah +1", http.StatusOK)
}

func (h *AIHandler) GetDailyQuizStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.service.GetDailyQuizStatus(userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil status daily quiz", http.StatusOK)
}

func (h *AIHandler) UpdateMaterial(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req dto.UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	resp, err := h.service.UpdateMaterial(id, userID.(string), role.(string), req)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Success(c.Writer, resp, "Materi berhasil diperbarui", http.StatusOK)
}

func (h *AIHandler) DeleteMaterial(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	err := h.service.DeleteMaterial(id, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Success(c.Writer, nil, "Materi berhasil dihapus", http.StatusOK)
}
