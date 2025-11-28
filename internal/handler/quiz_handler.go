package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type QuizHandler struct {
	service *service.QuizService
}

func NewQuizHandler(s *service.QuizService) *QuizHandler {
	return &QuizHandler{service: s}
}

func (h *QuizHandler) GetQuizzesByMaterial(c *gin.Context) {
	materialID := c.Param("id")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	resp, err := h.service.GetQuizzesByMaterial(materialID, userID.(string), role.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil daftar quiz", http.StatusOK)
}

func (h *QuizHandler) SubmitQuiz(c *gin.Context) {
	quizID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req dto.SubmitQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Format data jawaban salah", http.StatusBadRequest)
		return
	}

	resp, err := h.service.SubmitQuiz(quizID, userID.(string), req)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Quiz berhasil disubmit", http.StatusOK)
}

func (h *QuizHandler) GetAttemptReview(c *gin.Context) {
	quizId := c.Param("id")
	userID, _ := c.Get("user_id")

	resp, err := h.service.GetAttemptReview(quizId, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil detail hasil quiz", http.StatusOK)
}
