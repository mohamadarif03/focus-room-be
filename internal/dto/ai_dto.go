package dto

import "time"

type MaterialResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	SourceType   string    `json:"source_type"`
	Source       string    `json:"source"`
	Summary      string    `json:"summary"`
	PackageTitle string    `json:"package_title"`
	IsPublic     bool      `json:"is_public"`
	CreatedAt    time.Time `json:"created_at"`
}

type IngestYouTubeRequest struct {
	Title     string `json:"title"`
	URL       string `json:"url" binding:"required,url"`
	PackageID *uint  `json:"package_id"`
}

type GenerateSummaryRequest struct {
	MaterialID uint `json:"material_id" binding:"required"`
}

type GenerateSummaryResponse struct {
	MaterialID uint   `json:"material_id"`
	Summary    string `json:"summary"`
}

type GenerateQuizRequest struct {
	MaterialID    uint `json:"material_id" binding:"required"`
	QuestionCount int  `json:"question_count" binding:"required,min=1,max=10"`
}

type QuizOption map[string]string

type QuizQuestion struct {
	ID           int          `json:"id"`
	Pertanyaan   string       `json:"pertanyaan"`
	Pilihan      []QuizOption `json:"pilihan"`
	JawabanBenar string       `json:"jawaban_benar"`
}

type GenerateQuizResponse struct {
	MaterialID uint           `json:"material_id"`
	Questions  []QuizQuestion `json:"questions"`
}

type DailyQuizResponse struct {
	Questions interface{} `json:"questions"`
	IsDone    bool        `json:"is_done"`
}

type GenerateFlashcardRequest struct {
	MaterialID uint `json:"material_id" binding:"required"`
}
type ClaimQuizRequest struct {
	Score int `json:"score" binding:"required,min=10"`
}

type FlashcardItem struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

type GenerateFlashcardResponse struct {
	MaterialID uint            `json:"material_id"`
	Flashcards []FlashcardItem `json:"flashcards"`
}
