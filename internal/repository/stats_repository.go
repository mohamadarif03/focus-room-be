package repository

import (
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) CreateFocusSession(session *model.FocusSession) error {
	return r.db.Create(session).Error
}

func (r *StatsRepository) UpdateFocusDuration(sessionID, userID uint, duration int) error {
	result := r.db.Model(&model.FocusSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("duration", duration)

	return result.Error
}

func (r *StatsRepository) CreateQuizLog(log *model.QuizLog) error {
	return r.db.Create(log).Error
}

func (r *StatsRepository) SumStudyMinutes(userID uint, startDate, endDate *time.Time) (int64, error) {
	var totalMinutes int64
	query := r.db.Model(&model.FocusSession{}).Where("user_id = ?", userID)

	if startDate != nil && endDate != nil {
		query = query.Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	}

	err := query.Select("COALESCE(SUM(duration), 0)").Scan(&totalMinutes).Error
	return totalMinutes, err
}

func (r *StatsRepository) CountCompletedTasks(userID uint, startDate, endDate *time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&model.Task{}).Where("user_id = ? AND completed = ?", userID, true)

	if startDate != nil && endDate != nil {
		query = query.Where("task_date >= ? AND task_date <= ?", startDate, endDate)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *StatsRepository) CountQuizzesTaken(userID uint, startDate, endDate *time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&model.QuizLog{}).Where("user_id = ?", userID)

	if startDate != nil && endDate != nil {
		query = query.Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *StatsRepository) GetMostProductiveDay(userID uint, startDate, endDate *time.Time) (string, error) {
	type Result struct {
		DayName string
		Total   int
	}
	var result Result

	queryStr := `
		SELECT TRIM(TO_CHAR(task_date, 'Day')) as day_name, COUNT(*) as total
		FROM tasks
		WHERE user_id = ? AND completed = true
	`
	args := []interface{}{userID}

	if startDate != nil && endDate != nil {
		queryStr += " AND task_date >= ? AND task_date <= ?"
		args = append(args, startDate, endDate)
	}

	queryStr += " GROUP BY day_name ORDER BY total DESC LIMIT 1"

	err := r.db.Raw(queryStr, args...).Scan(&result).Error
	if err != nil {
		return "-", err
	}
	if result.DayName == "" {
		return "-", nil
	}
	return result.DayName, nil
}
