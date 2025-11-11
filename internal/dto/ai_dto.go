package dto

type MaterialResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
}

type IngestYouTubeRequest struct {
	Title string `json:"title" binding:"required"`
	URL   string `json:"url" binding:"required,url"`
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
