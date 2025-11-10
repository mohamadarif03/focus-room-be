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



func (s *TaskService) GetTasks(userIDString string, dateQuery string) ([]dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	var targetDate time.Time

	if dateQuery == "" {
		now := time.Now()
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		parsedDate, err := time.Parse("2006-01-02", dateQuery)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
		}
		targetDate = parsedDate
	}

	tasks, err := s.taskRepo.FindTasksByUserIDAndDate(uint(userID), targetDate)
	if err != nil {
		return nil, errors.New("gagal mengambil data task")
	}

	var taskResponses []dto.TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, dto.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			IsCompleted: task.IsCompleted,
			TaskDate:    task.TaskDate,
			UserID:      task.UserID,
		})
	}

	return taskResponses, nil
}
