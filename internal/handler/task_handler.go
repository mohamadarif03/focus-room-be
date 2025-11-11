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

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest

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

	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID dari token", http.StatusInternalServerError)
		return
	}

	response, err := h.service.CreateTask(req, userIDString.(string))
	if err != nil {
		if err.Error() == "format tanggal tidak valid, gunakan YYYY-MM-DD" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Task berhasil ditambahkan", http.StatusCreated)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	// 1. Ambil ID dari URL
	taskIDString := c.Param("id")

	// 2. Ambil UserID dari Context (WAJIB)
	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID dari token", http.StatusInternalServerError)
		return
	}

	// 3. Bind JSON
	var req dto.UpdateTaskRequest
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

	// 4. Panggil Service
	response, err := h.service.UpdateTask(taskIDString, userIDString.(string), req)
	if err != nil {
		if err.Error() == "task tidak ditemukan" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound) // 404
			return
		}
		if err.Error() == "akses ditolak: anda bukan pemilik task ini" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusForbidden) // 403
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, response, "Task berhasil diupdate", http.StatusOK)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskIDString := c.Param("id")

	userIDString, exists := c.Get("user_id")
	if !exists {
		utils.Error(c.Writer, nil, "Gagal mendapatkan user ID dari token", http.StatusInternalServerError)
		return
	}

	err := h.service.DeleteTask(taskIDString, userIDString.(string))
	if err != nil {
		if err.Error() == "task tidak ditemukan" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "akses ditolak: anda bukan pemilik task ini" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusForbidden)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Success(c.Writer, nil, "Task berhasil dihapus", http.StatusOK)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userIDString, _ := c.Get("user_id")

	dateQuery := c.Query("date")

	response, err := h.service.GetTasks(userIDString.(string), dateQuery)
	if err != nil {
		if err.Error() == "format tanggal tidak valid, gunakan YYYY-MM-DD" {
			utils.Error(c.Writer, nil, err.Error(), http.StatusBadRequest)
			return
		}
		utils.Error(c.Writer, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	message := "Berhasil mengambil tasks untuk hari ini"
	if dateQuery != "" {
		message = "Berhasil mengambil tasks untuk tanggal " + dateQuery
	}
	utils.Success(c.Writer, response, message, http.StatusOK)
}
