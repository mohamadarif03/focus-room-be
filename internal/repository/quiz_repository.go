package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type QuizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

func (r *QuizRepository) CreateWithLimit(quiz *model.Quiz) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Quiz{}).Where("material_id = ?", quiz.MaterialID).Count(&count).Error; err != nil {
			return err
		}
		if count >= 3 {
			var oldestQuiz model.Quiz
			if err := tx.Where("material_id = ?", quiz.MaterialID).Order("created_at asc").First(&oldestQuiz).Error; err != nil {
				return err
			}
			if err := tx.Delete(&oldestQuiz).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(quiz).Error; err != nil {
			return err
		}
		return nil
	})
}


func (r *QuizRepository) FindAttemptByID(attemptID uint) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt

	err := r.db.
		Preload("Answers").
		Preload("Quiz.Questions").
		Preload("Quiz.Material").
		Where("id = ?", attemptID).
		First(&attempt).Error

	return &attempt, err
}

func (r *QuizRepository) FindAttemptsByQuizID(quizID, userID uint) ([]model.QuizAttempt, error) {
	var attempts []model.QuizAttempt
	err := r.db.Where("quiz_id = ? AND user_id = ?", quizID, userID).
		Order("created_at desc").
		Find(&attempts).Error
	return attempts, err
}

func (r *QuizRepository) FindAllByMaterialID(materialID uint) ([]model.Quiz, error) {
	var quizzes []model.Quiz
	err := r.db.Preload("Questions").Where("material_id = ?", materialID).Order("created_at desc").Find(&quizzes).Error
	return quizzes, err
}

func (r *QuizRepository) FindByID(id uint) (*model.Quiz, error) {
	var quiz model.Quiz
	err := r.db.Preload("Questions").Where("id = ?", id).First(&quiz).Error
	return &quiz, err
}

func (r *QuizRepository) SaveAttempt(attempt *model.QuizAttempt) error {
	return r.db.Create(attempt).Error
}
