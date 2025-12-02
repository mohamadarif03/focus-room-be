package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(s *service.StatsService) *StatsHandler {
	return &StatsHandler{service: s}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	filter := c.DefaultQuery("filter", "all")

	resp, err := h.service.GetUserStats(userID.(string), filter)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Berhasil mengambil statistik", http.StatusOK)
}

func (h *StatsHandler) LogFocus(c *gin.Context) {
	var req dto.LogFocusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")
	err := h.service.LogFocusSession(userID.(string), req)
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, nil, "Sesi fokus berhasil dicatat", http.StatusCreated)
}

func (h *StatsHandler) StartFocus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.service.StartFocusSession(userID.(string))
	if err != nil {
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, resp, "Sesi fokus dimulai", http.StatusCreated)
}

func (h *StatsHandler) UpdateFocus(c *gin.Context) {
	var req dto.UpdateFocusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, nil, "Data tidak valid", http.StatusBadRequest)
		return
	}

	userID, _ := c.Get("user_id")

	err := h.service.UpdateFocusSession(userID.(string), req)
	if err != nil {
		utils.Error(c.Writer, nil, "Gagal mengupdate sesi (Mungkin sesi tidak ditemukan)", http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, nil, "Sesi berhasil diupdate", http.StatusOK)
}
