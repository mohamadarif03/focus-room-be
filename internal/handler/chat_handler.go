package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type ChatHandler struct {
	service *service.ChatService
}

func NewChatHandler(s *service.ChatService) *ChatHandler {
	return &ChatHandler{service: s}
}

// POST /chat
func (h *ChatHandler) SendChat(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.service.SendChat(c.Request.Context(), req, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Pesan terkirim", http.StatusOK)
}

// GET /chat?material_id=1
func (h *ChatHandler) GetHistory(c *gin.Context) {
	materialID := c.Query("material_id")
	userID, _ := c.Get("user_id")

	if materialID == "" {
		utils.Error(c.Writer, nil, "material_id wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.service.GetHistory(materialID, userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusForbidden)
		return
	}

	utils.Success(c.Writer, resp, "Riwayat chat berhasil diambil", http.StatusOK)
}
