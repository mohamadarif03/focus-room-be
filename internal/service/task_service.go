package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) CreateTask(req dto.CreateTaskRequest, userIDString string) (*dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	taskDate, err := time.Parse("2006-01-02", req.TaskDate)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}

	newTask := model.Task{
		Title:       req.Title,
		TaskDate:    taskDate,
		IsCompleted: false,
		UserID:      uint(userID), 
	}

	createdTask, err := s.taskRepo.CreateTask(&newTask)
	if err != nil {
		return nil, errors.New("gagal menyimpan task ke database")
	}

	response := &dto.TaskResponse{
		ID:          createdTask.ID,
		Title:       createdTask.Title,
		IsCompleted: createdTask.IsCompleted,
		TaskDate:    createdTask.TaskDate,
		UserID:      createdTask.UserID,
	}

	return response, nil
}