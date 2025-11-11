package repository

import (
	"time"

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

func (r *TaskRepository) FindTasksByUserIDAndDate(userID uint, date time.Time) ([]model.Task, error) {
	var tasks []model.Task

	err := r.db.Where("user_id = ? AND task_date = ?", userID, date).Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) FindByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.Where("id = ?", id).First(&task).Error

	if err != nil {
		return nil, err
	}

	return &task, err
}

func (r *TaskRepository) Update(task *model.Task) (*model.Task, error) {
	err := r.db.Save(&task).Error
	return task, err
}

func (r *TaskRepository) Delete(id uint) error {
	err := r.db.Unscoped().Where("id = ?", id).Delete(&model.Task{}).Error
	return err
}
