package dto

import "time"

type CreateTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	Context  string `json:"context"`
	TaskDate string `json:"task_date" binding:"required"`
	Priority string `json:"priority" binding:"required,oneof=low medium high"`
}

type UpdateTaskRequest struct {
	Title     string `json:"title" binding:"required"`
	Context   string `json:"context"`
	Priority  string `json:"priority" binding:"required,oneof=low medium high"`
	Completed bool   `json:"completed"`
}

type TaskResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Context   string    `json:"context"`
	Priority  string    `json:"priority"`
	TaskDate  time.Time `json:"task_date"`
	Completed bool      `json:"completed"`
	UserID    uint      `json:"user_id"`
}
