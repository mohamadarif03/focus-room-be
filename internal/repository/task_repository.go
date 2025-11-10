package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(task *model.Task) (*model.Task, error) {
	err := r.db.Create(&task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
