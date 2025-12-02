package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"gorm.io/gorm"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
	userRepo *repository.UserRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, userRepo *repository.UserRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}

func taskToResponse(task *model.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:        task.ID,
		Title:     task.Title,
		Context:   task.Context,
		Priority:  task.Priority,
		StartDate: task.StartDate,
		TaskDate:  task.TaskDate,
		Completed: task.Completed,
		UserID:    task.UserID,
	}
}

func (s *TaskService) CreateTask(req dto.CreateTaskRequest, userIDString string) (*dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	taskDate, err := time.Parse(time.RFC3339, req.TaskDate)
	if err != nil {
		return nil, errors.New("format tanggal task_date tidak valid (ISO 8601)")
	}

	var startDate *time.Time
	if req.StartDate != "" {
		parsedStart, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, errors.New("format tanggal start_date tidak valid (ISO 8601)")
		}
		startDate = &parsedStart
	}

	newTask := model.Task{
		Title:     req.Title,
		Context:   req.Context,
		StartDate: startDate,
		TaskDate:  taskDate,
		Priority:  req.Priority,
		Completed: false,
		UserID:    uint(userID),
	}

	createdTask, err := s.taskRepo.CreateTask(&newTask)
	if err != nil {
		return nil, errors.New("gagal menyimpan task ke database")
	}

	response := taskToResponse(createdTask)
	return &response, nil
}

func (s *TaskService) GetTasks(userIDString, dateQuery, priorityQuery string) ([]dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	var filterDate *time.Time

	if dateQuery != "" {
		parsedDate, err := time.Parse("2006-01-02", dateQuery)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
		}
		filterDate = &parsedDate
	}

	tasks, err := s.taskRepo.FindAllTasks(uint(userID), filterDate, priorityQuery)
	if err != nil {
		return nil, errors.New("gagal mengambil data task")
	}

	var taskResponses []dto.TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, taskToResponse(&task))
	}

	return taskResponses, nil
}

func (s *TaskService) UpdateTask(taskIDString string, userIDString string, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}
	taskID, err := strconv.ParseUint(taskIDString, 10, 32)

	if err != nil {
		return nil, errors.New("task ID tidak valid")
	}

	task, err := s.taskRepo.FindTaskByID(uint(taskID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data task")
	}

	if task.UserID != uint(userID) {
		return nil, errors.New("akses ditolak: anda bukan pemilik task ini")
	}

	task.Title = req.Title
	task.Context = req.Context
	task.Priority = req.Priority
	task.Completed = req.Completed
	if req.StartDate != "" {
		parsedStart, err := time.Parse(time.RFC3339, req.StartDate)
		if err == nil {
			task.StartDate = &parsedStart
		}
	} 

	updatedTask, err := s.taskRepo.UpdateTask(task)
	if err != nil {
		return nil, errors.New("gagal mengupdate task")
	}

	response := taskToResponse(updatedTask)
	return &response, nil
}

func (s *TaskService) DeleteTask(taskIDString string, userIDString string) error {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return errors.New("user ID tidak valid")
	}
	taskID, err := strconv.ParseUint(taskIDString, 10, 32)
	if err != nil {
		return errors.New("task ID tidak valid")
	}

	task, err := s.taskRepo.FindTaskByID(uint(taskID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task tidak ditemukan")
		}
		return errors.New("gagal mengambil data task")
	}

	if task.UserID != uint(userID) {
		return errors.New("akses ditolak: anda bukan pemilik task ini")
	}

	err = s.taskRepo.Delete(uint(taskID))
	if err != nil {
		return errors.New("gagal menghapus task")
	}

	return nil
}
