package handler

import (
	"net/http"

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

	title := c.PostForm("title")
	if title == "" {
		utils.Error(c.Writer, nil, "Judul materi wajib diisi", http.StatusBadRequest)
		return
	}
	pkgID := c.PostForm("package_id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role") // Ambil Role untuk penanda IsPublic

	resp, err := h.service.IngestPDF(c.Request.Context(), fileHeader, title, pkgID, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil memproses PDF", http.StatusCreated)
}

// POST /materials/youtube
func (h *AIHandler) IngestYouTube(c *gin.Context) {
	var req dto.IngestYouTubeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role") // Ambil Role untuk penanda IsPublic

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

// POST /ai/quiz
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

// GET /daily-quiz
func (h *AIHandler) GetDailyQuiz(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.service.GetDailyQuiz(c.Request.Context(), userID.(string))
	if err != nil {
		if err.Error() == "anda belum mengupload materi apapun" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil Daily Quiz", http.StatusOK)
}

// POST /daily-quiz/claim
func (h *AIHandler) ClaimDailyStreak(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.ClaimQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	// Validasi Logic: Backend hanya menerima jika score == 10 (Benar Semua)
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