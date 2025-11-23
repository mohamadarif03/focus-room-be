package dto

import "time"

type QuizSimpleResponse struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	QuestionCount int       `json:"question_count"`
	Score         *int      `json:"score"`
}

type SubmitQuizRequest struct {
	Answers []QuizAnswerItem `json:"answers" binding:"required,dive"`
}

type QuizAnswerItem struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	UserAnswer string `json:"user_answer" binding:"required"`
}

type SubmitQuizResponse struct {
	AttemptID uint `json:"attempt_id"`
	Score     int  `json:"score"`
	Correct   int  `json:"correct"`
	Wrong     int  `json:"wrong"`
}

type QuizReviewItem struct {
	ID            uint         `json:"id"`
	Pertanyaan    string       `json:"pertanyaan"`
	Pilihan       []OptionItem `json:"pilihan"`
	UserAnswer    string       `json:"user_answer"`
	CorrectAnswer string       `json:"correct_answer"`
	IsCorrect     bool         `json:"is_correct"`
}

type QuizAttemptReviewResponse struct {
	AttemptID uint             `json:"id"`
	QuizID    uint             `json:"quiz_id"`
	Score     int              `json:"score"`
	CreatedAt time.Time        `json:"created_at"`
	Questions []QuizReviewItem `json:"questions"`
}
