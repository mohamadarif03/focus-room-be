package service

import (
	"errors"
	"strconv"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
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
