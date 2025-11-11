package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type AIHandler struct {
	service *service.AIService
}

func NewAIHandler(s *service.AIService) *AIHandler {
	return &AIHandler{service: s}
}

func (h *AIHandler) SummarizePDF(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		utils.Error(c.Writer, nil, "File 'pdf' tidak ditemukan di request", http.StatusBadRequest)
		return
	}

	if file.Header.Get("Content-Type") != "application/pdf" {
		utils.Error(c.Writer, nil, "File harus berekstensi .pdf", http.StatusBadRequest)
		return
	}

	response, err := h.service.SummarizePDF(c.Request.Context(), file)
	if err != nil {
		if err.Error() == "PDF ini tidak mengandung teks" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "PDF berhasil dirangkum", http.StatusOK)
}
