package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

// Mencatat sesi fokus selesai
func (s *StatsService) LogFocusSession(userIDString string, req dto.LogFocusRequest) error {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	session := &model.FocusSession{
		UserID:    uint(userID),
		Duration:  req.Duration,
		CreatedAt: time.Now(),
	}
	return s.statsRepo.CreateFocusSession(session)
}

func (s *StatsService) GetUserStats(userIDString, filter string) (*dto.StatsResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}
	uID := uint(userID)

	var startDate, endDate *time.Time
	now := time.Now()

	switch filter {
	case "day":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		end := start.Add(24 * time.Hour)
		startDate, endDate = &start, &end
	case "week":
		start := now.AddDate(0, 0, -7)
		startDate, endDate = &start, &now
	case "month":
		start := now.AddDate(0, -1, 0)
		startDate, endDate = &start, &now
	case "year":
		start := now.AddDate(-1, 0, 0)
		startDate, endDate = &start, &now
	default:
		startDate, endDate = nil, nil
	}

	totalMinutes, err := s.statsRepo.SumStudyMinutes(uID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	tasksCompleted, err := s.statsRepo.CountCompletedTasks(uID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	quizzesTaken, err := s.statsRepo.CountQuizzesTaken(uID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	mostProdDay, err := s.statsRepo.GetMostProductiveDay(uID, startDate, endDate)
	if err != nil {
		mostProdDay = "-"
	}

	// Format Menit ke Jam
	hours := totalMinutes / 60
	mins := totalMinutes % 60
	timeString := fmt.Sprintf("%dh", hours)
	if mins > 0 {
		timeString = fmt.Sprintf("%dh %dm", hours, mins)
	}

	// Translate Hari
	dayMap := map[string]string{
		"Monday": "Senin", "Tuesday": "Selasa", "Wednesday": "Rabu",
		"Thursday": "Kamis", "Friday": "Jumat", "Saturday": "Sabtu", "Sunday": "Minggu",
	}
	if val, ok := dayMap[mostProdDay]; ok {
		mostProdDay = val
	}

	return &dto.StatsResponse{
		TotalStudyHours:   timeString,
		TasksCompleted:    tasksCompleted,
		QuizzesTaken:      quizzesTaken,
		MostProductiveDay: mostProdDay,
	}, nil
}

func (s *StatsService) StartFocusSession(userIDString string) (*dto.StartFocusResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	session := &model.FocusSession{
		UserID:    uint(userID),
		Duration:  0,
		CreatedAt: time.Now(),
	}

	if err := s.statsRepo.CreateFocusSession(session); err != nil {
		return nil, err
	}

	return &dto.StartFocusResponse{
		SessionID: session.ID,
	}, nil
}

func (s *StatsService) UpdateFocusSession(userIDString string, req dto.UpdateFocusRequest) error {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	return s.statsRepo.UpdateFocusDuration(req.SessionID, uint(userID), req.Duration)
}
