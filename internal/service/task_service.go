package service

import (
	"errors"
	"log"
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

// taskToResponse adalah helper untuk konversi model Task ke DTO TaskResponse
func taskToResponse(task *model.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:        task.ID,
		Title:     task.Title,
		Context:   task.Context,
		Priority:  task.Priority,
		TaskDate:  task.TaskDate,
		Completed: task.Completed,
		UserID:    task.UserID,
	}
}

func (s *TaskService) CreateTask(req dto.CreateTaskRequest, userIDString string) (*dto.TaskResponse, error) {
	// 1. Konversi UserID
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	// 2. Parsing string ISO 8601 (RFC3339)
	taskDate, err := time.Parse(time.RFC3339, req.TaskDate)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan format ISO 8601 (RFC3339)")
	}

	// 3. Buat model Task baru
	newTask := model.Task{
		Title:     req.Title,
		Context:   req.Context,
		TaskDate:  taskDate,
		Priority:  req.Priority,
		Completed: false, // Task baru selalu 'false'
		UserID:    uint(userID),
	}

	// 4. Simpan ke database
	createdTask, err := s.taskRepo.CreateTask(&newTask)
	if err != nil {
		return nil, errors.New("gagal menyimpan task ke database")
	}

	// 5. Kembalikan sebagai DTO Response
	response := taskToResponse(createdTask)
	return &response, nil
}

func (s *TaskService) GetTasks(userIDString string, dateQuery string) ([]dto.TaskResponse, error) {
	// 1. Konversi UserID
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	var targetDate time.Time

	// 2. Logika Tanggal
	if dateQuery == "" {
		// Jika query ?date= kosong, pakai tanggal hari ini
		now := time.Now()
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		// Jika query ?date= ada, parse tanggal YYYY-MM-DD
		parsedDate, err := time.Parse("2006-01-02", dateQuery)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
		}
		targetDate = parsedDate
	}

	// 3. Panggil Repository
	// Repositori (FindTasksByUserIDAndDate) sudah di-update untuk
	// mencari berdasarkan rentang 24 jam (>= date AND < date+24jam)
	tasks, err := s.taskRepo.FindTasksByUserIDAndDate(uint(userID), targetDate)
	if err != nil {
		return nil, errors.New("gagal mengambil data task")
	}

	// 4. Konversi Model ke DTO
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

	updatedTask, err := s.taskRepo.UpdateTask(task)
	if err != nil {
		return nil, errors.New("gagal mengupdate task")
	}

	if req.Completed {
		go s.checkRealtimeStreak(uint(userID), task.TaskDate)
	}

	response := taskToResponse(updatedTask)
	return &response, nil
}

func (s *TaskService) checkRealtimeStreak(userID uint, taskDate time.Time) {
	log.Printf("[Streak H-0] User %d menyelesaikan task. Memeriksa...", userID)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	taskDay := time.Date(taskDate.Year(), taskDate.Month(), taskDate.Day(), 0, 0, 0, 0, time.Local)
	if !taskDay.Equal(today) {
		log.Printf("[Streak H-0] User %d menyelesaikan task lama. Streak H-0 diabaikan.", userID)
		return
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return
	}

	if user.LastStreakAwardedDate != nil {
		lastAward := *user.LastStreakAwardedDate
		lastAwardDate := time.Date(lastAward.Year(), lastAward.Month(), lastAward.Day(), 0, 0, 0, 0, time.Local)
		if lastAwardDate.Equal(today) {
			log.Printf("[Streak H-0] User %d sudah dapat imbalan hari ini. Diabaikan.", userID)
			return
		}
	}

	tasksToday, err := s.taskRepo.FindTasksByUserIDAndDate(userID, today)
	if err != nil {
		return
	}

	totalTasks := len(tasksToday)
	completedTasks := 0
	for _, t := range tasksToday {
		if t.Completed {
			completedTasks++
		}
	}

	if totalTasks > 0 && totalTasks == completedTasks {
		user.CurrentStreak += 1
		user.LastStreakAwardedDate = &now

		_, err := s.userRepo.Update(user)
		if err == nil {
			log.Printf("[Streak H-0] SUKSES! User %d menyelesaikan semua task H-0. Streak naik ke %d.", userID, user.CurrentStreak)
		}
	} else {
		log.Printf("[Streak H-0] User %d belum selesai. (Total: %d, Selesai: %d)", userID, totalTasks, completedTasks)
	}
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