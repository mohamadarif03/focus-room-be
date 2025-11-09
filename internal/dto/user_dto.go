package dto

import "time"

type UserResponse struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	CurrentStreak int       `json:"current_streak"`
	CreatedAt     time.Time `json:"created_at"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=siswa pembimbing admin"`
}
