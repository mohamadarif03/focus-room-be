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

	verificationToken := uuid.New().String()

	newUser := &model.User{
		Username:          req.Username,
		Email:             req.Email,
		PasswordHash:      hashedPassword,
		Role:              req.Role,
		IsVerified:        false,
		VerificationToken: verificationToken,
		CreatedAt:         time.Now(),
	}

	createdUser, err := s.userRepo.CreateUser(newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	go func() {
		if err := utils.SendVerificationEmail(createdUser.Email, verificationToken); err != nil {
			fmt.Printf("Gagal mengirim email verifikasi ke %s: %v\n", createdUser.Email, err)
		}
	}()

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

	if !user.IsVerified {
		return nil, errors.New("email belum diverifikasi. silakan cek kotak masuk anda")
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

func (s *AuthService) VerifyEmail(token string) error {
	var user model.User
	err := s.userRepo.DB.Where("verification_token = ?", token).First(&user).Error
	if err != nil {
		return errors.New("token verifikasi tidak valid")
	}

	if user.IsVerified {
		return errors.New("email sudah terverifikasi")
	}

	user.IsVerified = true
	user.VerificationToken = ""

	if _, err := s.userRepo.Update(&user); err != nil { // Assuming Update accepts *model.User
		return errors.New("gagal memverifikasi user")
	}

	return nil
}

func (s *AuthService) ResendVerificationCode(req dto.ResendVerificationRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("email tidak terdaftar")
		}
		return errors.New("database error")
	}

	if user.IsVerified {
		return errors.New("email sudah terverifikasi")
	}

	// Generate new verification token
	verificationToken := uuid.New().String()
	user.VerificationToken = verificationToken

	if _, err := s.userRepo.Update(user); err != nil {
		return errors.New("gagal mengupdate token verifikasi")
	}

	go func() {
		if err := utils.SendVerificationEmail(user.Email, verificationToken); err != nil {
			fmt.Printf("Gagal mengirim ulang email verifikasi ke %s: %v\n", user.Email, err)
		}
	}()

	return nil
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

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("email tidak ditemukan")
		}
		return errors.New("database error")
	}

	// Generate Reset Token
	resetToken := uuid.New().String()
	expiry := time.Now().Add(1 * time.Hour)

	user.ResetPasswordToken = resetToken
	user.ResetTokenExpiry = &expiry

	if _, err := s.userRepo.Update(user); err != nil {
		return errors.New("gagal menyimpan token reset password")
	}

	// Send Email Async
	go func() {
		if err := utils.SendResetPasswordEmail(user.Email, resetToken); err != nil {
			fmt.Printf("Gagal mengirim email reset password ke %s: %v\n", user.Email, err)
		}
	}()

	return nil
}

func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	var user model.User
	if err := s.userRepo.DB.Where("reset_password_token = ?", req.Token).First(&user).Error; err != nil {
		return errors.New("token tidak valid atau kadaluarsa")
	}

	if user.ResetTokenExpiry != nil && user.ResetTokenExpiry.Before(time.Now()) {
		return errors.New("token sudah kadaluarsa")
	}

	if req.Password != req.PasswordConfirm {
		return errors.New("password dan konfirmasi password tidak cocok")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("gagal mengamankan password")
	}

	user.PasswordHash = hashedPassword
	user.ResetPasswordToken = ""
	user.ResetTokenExpiry = nil

	if _, err := s.userRepo.Update(&user); err != nil {
		return errors.New("gagal mereset password")
	}

	return nil
}
