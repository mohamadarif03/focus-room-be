package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/google/generative-ai-go/genai"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
	"google.golang.org/api/option"
)

type AIService struct {
	geminiModel *genai.GenerativeModel
}

func NewAIService(apiKey string) (*AIService, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY tidak ditemukan")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.5-flash")

	return &AIService{geminiModel: model}, nil
}

func (s *AIService) SummarizePDF(ctx context.Context, fileHeader *multipart.FileHeader) (*dto.SummaryResponse, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("gagal membuka file")
	}
	defer file.Close()

	log.Println("Mulai mengekstrak teks dari PDF...")
	rawText, err := utils.ExtractTextFromPDF(file, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("gagal ekstrak PDF: %w", err)
	}

	if rawText == "" {
		return nil, errors.New("PDF ini tidak mengandung teks")
	}
	log.Println("Ekstrak teks selesai. Jumlah karakter:", len(rawText))

	prompt := fmt.Sprintf("Kamu adalah asisten belajar yang ahli. Tolong rangkum materi dari teks PDF berikut ini secara jelas, rinci, dan terstruktur. Gunakan poin-poin jika perlu. Berikut adalah teksnya:\n\n---\n\n%s", rawText)

	log.Println("Mengirim request ke Gemini...")
	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal memanggil Gemini: %w", err)
	}

	var summary string
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				summary += string(txt)
			}
		}
	}

	if summary == "" {
		return nil, errors.New("Gemini tidak memberikan rangkuman")
	}
	log.Println("Rangkuman diterima.")

	response := &dto.SummaryResponse{
		Summary:  summary,
		Filename: fileHeader.Filename,
	}

	return response, nil
}
