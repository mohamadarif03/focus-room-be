package dto

type StatsResponse struct {
	TotalStudyHours   string `json:"total_study_hours"`
	TasksCompleted    int64  `json:"tasks_completed"`
	QuizzesTaken      int64  `json:"quizzes_taken"`
	MostProductiveDay string `json:"most_productive_day"`
}

type LogFocusRequest struct {
	Duration int `json:"duration" binding:"required,min=1"`
}
