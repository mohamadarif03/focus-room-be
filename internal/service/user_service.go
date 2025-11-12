package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
	taskRepo *repository.TaskRepository
}

func NewUserService(userRepo *repository.UserRepository, taskRepo *repository.TaskRepository) *UserService {
	return &UserService{userRepo: userRepo, taskRepo: taskRepo}
}

func (s *UserService) GetAllUsers() ([]dto.UserResponse, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, userToResponse(&user))
	}
	return userResponses, nil
}

func (s *UserService) GetSelf(userIDString string) (*dto.UserResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	response := userToResponse(user)
	return &response, nil
}

func (s *UserService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	response := userToResponse(user)
	return &response, nil
}

func (s *UserService) UpdateUser(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if user.Email != req.Email {
		existing, err := s.userRepo.FindByEmail(req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("database error")
		}
		if existing != nil {
			return nil, errors.New("email sudah terdaftar")
		}
	}

	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role

	updatedUser, err := s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	response := userToResponse(updatedUser)
	return &response, nil
}

func (s *UserService) DeleteUser(id uint) error {
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	return s.userRepo.Delete(id)
}

func (s *UserService) CheckAndUpdateStreak(userIDString string) (*model.User, error) {
	log.Printf("[Streak H-1] Pengecekan 'Satpam' diminta oleh user: %s", userIDString)

	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if user.LastStreakCheckDate != nil {
		lastCheck := *user.LastStreakCheckDate
		lastCheckDate := time.Date(lastCheck.Year(), lastCheck.Month(), lastCheck.Day(), 0, 0, 0, 0, time.Local)

		if lastCheckDate.Equal(today) {
			log.Printf("[Streak H-1] User %d sudah dicek hari ini. Tidak ada update.", userID)
			return user, nil
		}
	}

	yesterday := today.Add(-24 * time.Hour)
	log.Printf("[Streak H-1] User %d belum dicek. Mengecek tugas H-1 (%s)...", userID, yesterday.Format("2006-01-02"))

	tasks, err := s.taskRepo.FindTasksByUserIDAndDate(user.ID, yesterday)
	if err != nil {
		return nil, fmt.Errorf("gagal ambil task user %d: %w", user.ID, err)
	}

	totalTasks := len(tasks)
	completedTasks := 0
	for _, task := range tasks {
		if task.IsCompleted {
			completedTasks++
		}
	}

	streakUpdated := false
	if totalTasks == 0 {
		if user.CurrentStreak > 0 {
			log.Printf("[Streak H-1] User %d: 0 tugas H-1. Streak reset ke 0.", user.ID)
			user.CurrentStreak = 0
			streakUpdated = true
		}
	} else if totalTasks > 0 && completedTasks != totalTasks {
		if user.CurrentStreak > 0 {
			log.Printf("[Streak H-1] User %d: Tugas H-1 tidak selesai. Streak reset ke 0.", user.ID)
			user.CurrentStreak = 0
			streakUpdated = true
		}
	} else {
		log.Printf("[Streak H-1] User %d: Sukses H-1. Tidak ada perubahan (imbalan sudah diberikan H-0).", user.ID)
	}
	// ------------------------------------------

	user.LastStreakCheckDate = &now

	_, err = s.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("gagal update streak user %d: %w", user.ID, err)
	}

	if streakUpdated {
		log.Printf("[Streak H-1] User %d berhasil di-reset.", user.ID)
	} else {
		log.Printf("[Streak H-1] Pengecekan User %d selesai, tidak ada reset.", user.ID)
	}

	return user, nil
}

func userToResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		CurrentStreak: user.CurrentStreak,
		CreatedAt:     user.CreatedAt,
	}
}
