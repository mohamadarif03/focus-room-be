package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
	"google.golang.org/api/option"
)

type AIService struct {
	geminiModel *genai.GenerativeModel
	matRepo     *repository.MaterialRepository
}

func NewAIService(apiKey string, matRepo *repository.MaterialRepository) (*AIService, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY tidak ditemukan")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.5-flash")
	return &AIService{geminiModel: model, matRepo: matRepo}, nil
}

func (s *AIService) IngestPDF(ctx context.Context, fileHeader *multipart.FileHeader, title string, userIDString string) (*dto.MaterialResponse, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("gagal membuka file")
	}
	defer file.Close()

	rawText, err := utils.ExtractTextFromPDF(file, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("gagal ekstrak PDF: %w", err)
	}
	if rawText == "" {
		return nil, errors.New("PDF ini tidak mengandung teks")
	}

	userID, _ := strconv.ParseUint(userIDString, 10, 32)
	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         title,
		SourceType:    "pdf",
		Source:        fileHeader.Filename,
		ExtractedText: rawText,
	}
	savedMat, err := s.matRepo.Save(newMaterial)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan materi: %w", err)
	}

	log.Println("PDF Ingested, ID:", savedMat.ID)
	return &dto.MaterialResponse{
		ID:         savedMat.ID,
		Title:      savedMat.Title,
		SourceType: savedMat.SourceType,
		Source:     savedMat.Source,
	}, nil
}

func (s *AIService) IngestYouTube(ctx context.Context, req dto.IngestYouTubeRequest, userIDString string) (*dto.MaterialResponse, error) {
	rawText, err := utils.ExtractTextFromYouTube(req.URL)
	if err != nil {
		return nil, err
	}
	if rawText == "" {
		return nil, errors.New("video ini tidak memiliki transkrip")
	}

	userID, _ := strconv.ParseUint(userIDString, 10, 32)
	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         req.Title,
		SourceType:    "youtube",
		Source:        req.URL,
		ExtractedText: rawText,
	}
	savedMat, err := s.matRepo.Save(newMaterial)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan materi: %w", err)
	}

	log.Println("YouTube Ingested, ID:", savedMat.ID)
	return &dto.MaterialResponse{
		ID:         savedMat.ID,
		Title:      savedMat.Title,
		SourceType: savedMat.SourceType,
		Source:     savedMat.Source,
	}, nil
}

func (s *AIService) GenerateSummary(ctx context.Context, req dto.GenerateSummaryRequest, userIDString string) (*dto.GenerateSummaryResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak punya akses")
	}

	prompt := fmt.Sprintf("Jelaskan ulang isi materi berikut secara jelas, mendalam, dan terstruktur, seperti seorang dosen profesional yang menjelaskan konsep di kelas, namun tanpa sapaan pembuka atau penutup kelas (misalnya: “Selamat pagi mahasiswa”, “Apakah ada pertanyaan?”, dan sejenisnya). Saat menjelaskan ulang: Gunakan bahasa yang natural, komunikatif, dan logis, bukan formal kaku. Fokus untuk memperjelas isi materi, bukan sekadar merangkum. Jelaskan konsep dan ide utama dengan contoh nyata atau analogi jika perlu. Jika ada istilah sulit, jelaskan maknanya terlebih dahulu sebelum lanjut. Gunakan gaya penjelasan yang mengalir seperti narasi dosen yang fokus menjelaskan isi (tanpa salam, tanpa tanya jawab). Tutup dengan ringkasan inti dan kesimpulan, bukan kalimat interaktif seperti “ada pertanyaan?” atau “sampai jumpa”. Output yang diharapkan: Penjelasan ulang yang runtut, detail, dan mudah dipahami Gaya profesional namun tetap natural Tidak ada bagian sapaan, humor, atau tanya-jawab interaktif. Materi:\n\n%s", material.ExtractedText)
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

	return &dto.GenerateSummaryResponse{
		MaterialID: req.MaterialID,
		Summary:    summary,
	}, nil
}

func (s *AIService) GenerateQuiz(ctx context.Context, req dto.GenerateQuizRequest, userIDString string) (*dto.GenerateQuizResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak punya akses")
	}

	prompt := fmt.Sprintf(
		`Buatkan %d soal latihan berdasarkan materi berikut.

Hasilkan dalam format JSON array yang valid dan rapi.
Setiap objek di dalam array harus memiliki struktur berikut:

{
  "id": number,
  "pertanyaan": "string",
  "pilihan": [
    {"A": "string"},
    {"B": "string"},
    {"C": "string"},
    {"D": "string"}
  ],
  "jawaban_benar": "string" // huruf A, B, C, atau D saja
}

Instruksi penting:
1. Soal harus relevan langsung dengan isi materi dan menguji pemahaman konsep (bukan hafalan).
2. Setiap opsi jawaban harus masuk akal dan proporsional, tidak terlalu mudah ditebak.
3. Hindari pola yang membuat jawaban benar selalu mudah dikenali, seperti:
   - jawaban paling panjang atau paling detail,
   - posisi jawaban benar selalu sama.
4. Gunakan bahasa Indonesia yang natural dan jelas, seperti soal buatan manusia.
5. Variasikan tingkat kesulitan: sebagian soal dasar, sebagian penerapan atau analisis.
6. Jangan tambahkan penjelasan, pembuka, atau teks apa pun di luar format JSON.
7. Pastikan output adalah JSON yang valid dan bisa langsung di-parse tanpa error.
8. Nilai "jawaban_benar" hanya berisi huruf A, B, C, atau D — bukan teks jawaban.

Materi:
%s`,
		req.QuestionCount,
		material.ExtractedText,
	)

	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal memanggil Gemini: %w", err)
	}

	var quizJSON string
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				quizJSON += string(txt)
			}
		}
	}

	quizJSON = strings.TrimSpace(quizJSON)
	quizJSON = strings.TrimPrefix(quizJSON, "```json")
	quizJSON = strings.TrimSuffix(quizJSON, "```")

	if quizJSON == "" {
		return nil, errors.New("Gemini tidak memberikan kuis")
	}

	var questions []dto.QuizQuestion
	if err := json.Unmarshal([]byte(quizJSON), &questions); err != nil {
		return nil, fmt.Errorf("gagal parsing hasil Gemini: %w", err)
	}

	return &dto.GenerateQuizResponse{
		MaterialID: req.MaterialID,
		Questions:  questions,
	}, nil
}
