package dto

import "time"

type CreateTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	TaskDate string `json:"task_date" binding:"required"` 
}

type TaskResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	IsCompleted bool      `json:"is_completed"`
	TaskDate    time.Time `json:"task_date"`
	UserID      uint      `json:"user_id"`
}