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

func (s *TaskService) UpdateTask(taskIDString string, userIDString string, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}
	taskID, err := strconv.ParseUint(taskIDString, 10, 32)
	if err != nil {
		return nil, errors.New("task ID tidak valid")
	}

	task, err := s.taskRepo.FindByID(uint(taskID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data task")
	}

	if task.UserID != uint(userID) {
		return nil, errors.New("akses ditolak: anda bukan pemilik task ini") // 403
	}

	task.Title = req.Title
	task.IsCompleted = req.IsCompleted

	updatedTask, err := s.taskRepo.Update(task)
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

	task, err := s.taskRepo.FindByID(uint(taskID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task tidak ditemukan")
		}
		return errors.New("gagal mengambil data task")
	}

	if task.UserID != uint(userID) {
		return errors.New("akses ditolak: anda bukan pemilik task ini") // 403
	}

	err = s.taskRepo.Delete(uint(taskID))
	if err != nil {
		return errors.New("gagal menghapus task")
	}

	return nil
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
		taskResponses = append(taskResponses, taskToResponse(&task))
	}

	return taskResponses, nil
}

func taskToResponse(task *model.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		IsCompleted: task.IsCompleted,
		TaskDate:    task.TaskDate,
		UserID:      task.UserID,
	}
}
