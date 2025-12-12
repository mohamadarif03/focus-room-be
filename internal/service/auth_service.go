package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("database error")
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	if req.Password != req.PasswordConfirm {
		return nil, errors.New("password and password_confirm doesn't match")
	}

	newUser := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
	}

	createdUser, err := s.userRepo.CreateUser(&newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	token, err := utils.GenerateToken(createdUser.ID, createdUser.Role)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	response := &dto.AuthResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Email:    createdUser.Email,
		Role:     createdUser.Role,
		Token:    token,
	}

	return response, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email atau password salah")
		}
		return nil, errors.New("database error")
	}

	isValidPassword := utils.VerifyPassword(user.PasswordHash, req.Password)
	if !isValidPassword {
		return nil, errors.New("email atau password salah")
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	response := &dto.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Token:    token,
	}

	return response, nil
}

func (s *AuthService) GoogleLogin(req dto.GoogleLoginRequest, ctx context.Context) (*dto.AuthResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("server misconfiguration: GOOGLE_CLIENT_ID not set")
	}

	payload, err := idtoken.Validate(ctx, req.Token, clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google token: %v", err)
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)

	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("database error")
	}

	var user *model.User

	if existingUser != nil {
		user = existingUser
	} else {
		randomPassword := uuid.New().String()
		hashedPassword, _ := utils.HashPassword(randomPassword)

		newUser := &model.User{
			Username:     name,
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         "siswa",
			CreatedAt:    time.Now(),
		}

		createdUser, err := s.userRepo.CreateUser(newUser)
		if err != nil {
			return nil, errors.New("failed to create user from google")
		}
		user = createdUser
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	response := &dto.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Token:    token,
	}

	return response, nil
}
