package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
)

type QuizService struct {
	quizRepo *repository.QuizRepository
	matRepo  *repository.MaterialRepository
}

func NewQuizService(quizRepo *repository.QuizRepository, matRepo *repository.MaterialRepository) *QuizService {
	return &QuizService{
		quizRepo: quizRepo,
		matRepo:  matRepo,
	}
}

func (s *QuizService) GetQuizzesByMaterial(materialIDString, userIDString, userRole string) ([]dto.QuizSimpleResponse, error) {
	materialID, _ := strconv.ParseUint(materialIDString, 10, 32)
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	_, err := s.matRepo.FindOneAccessible(uint(materialID), uint(userID), userRole)
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau akses ditolak")
	}

	quizzes, err := s.quizRepo.FindAllByMaterialID(uint(materialID))
	if err != nil {
		return nil, err
	}

	var response []dto.QuizSimpleResponse
	for _, q := range quizzes {

		score, _ := s.quizRepo.GetUserLatestScore(q.ID, uint(userID))

		response = append(response, dto.QuizSimpleResponse{
			ID:            q.ID,
			CreatedAt:     q.CreatedAt,
			QuestionCount: len(q.Questions),
			Score:         score,
		})
	}

	return response, nil
}

func (s *QuizService) SubmitQuiz(quizIDString, userIDString string, req dto.SubmitQuizRequest) (*dto.SubmitQuizResponse, error) {
	quizID, _ := strconv.ParseUint(quizIDString, 10, 32)
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	quiz, err := s.quizRepo.FindByID(uint(quizID))
	if err != nil {
		return nil, errors.New("quiz tidak ditemukan")
	}

	correctCount := 0
	var attemptDetails []model.QuizAttemptDetail

	answerKey := make(map[uint]string)
	for _, q := range quiz.Questions {
		answerKey[q.ID] = q.JawabanBenar
	}

	for _, ans := range req.Answers {
		correctAnswer, exists := answerKey[ans.QuestionID]
		if !exists {
			continue
		}

		isCorrect := (ans.UserAnswer == correctAnswer)
		if isCorrect {
			correctCount++
		}

		attemptDetails = append(attemptDetails, model.QuizAttemptDetail{
			QuestionID: ans.QuestionID,
			UserAnswer: ans.UserAnswer,
			IsCorrect:  isCorrect,
		})
	}

	totalQuestions := len(quiz.Questions)
	score := 0
	if totalQuestions > 0 {
		score = (correctCount * 100) / totalQuestions
	}

	attempt := model.QuizAttempt{
		UserID:    uint(userID),
		QuizID:    uint(quizID),
		Score:     score,
		CreatedAt: time.Now(),
		Answers:   attemptDetails,
	}

	if err := s.quizRepo.SaveAttempt(&attempt); err != nil {
		return nil, errors.New("gagal menyimpan hasil quiz")
	}

	return &dto.SubmitQuizResponse{
		AttemptID: attempt.ID,
		Score:     score,
		Correct:   correctCount,
		Wrong:     totalQuestions - correctCount,
	}, nil
}

func (s *QuizService) GetAttemptReview(quizIdString, userIDString string) (*dto.QuizAttemptReviewResponse, error) {
	quizId, _ := strconv.ParseUint(quizIdString, 10, 32)
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	attempt, err := s.quizRepo.FindAttemptByQuizID(uint(quizId))
	if err != nil {
		return nil, errors.New("riwayat quiz tidak ditemukan")
	}

	if attempt.UserID != uint(userID) {
		return nil, errors.New("akses ditolak")
	}

	var questionReviews []dto.QuizReviewItem

	userAnswersMap := make(map[uint]model.QuizAttemptDetail)
	for _, ans := range attempt.Answers {
		userAnswersMap[ans.QuestionID] = ans
	}

	for _, q := range attempt.Quiz.Questions {
		var pilihanItems []dto.OptionItem
		json.Unmarshal(q.Pilihan, &pilihanItems)

		userAnsDetail, answered := userAnswersMap[q.ID]
		userAnswer := ""
		isCorrect := false

		if answered {
			userAnswer = userAnsDetail.UserAnswer
			isCorrect = userAnsDetail.IsCorrect
		}

		questionReviews = append(questionReviews, dto.QuizReviewItem{
			ID:            q.ID,
			Pertanyaan:    q.Pertanyaan,
			Pilihan:       pilihanItems,
			UserAnswer:    userAnswer,
			CorrectAnswer: q.JawabanBenar,
			IsCorrect:     isCorrect,
		})
	}

	return &dto.QuizAttemptReviewResponse{
		AttemptID: attempt.ID,
		QuizID:    attempt.QuizID,
		Score:     attempt.Score,
		CreatedAt: attempt.CreatedAt,
		Questions: questionReviews,
	}, nil
}
