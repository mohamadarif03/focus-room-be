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
	"time"

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
	pkgRepo     *repository.PackageRepository
	userRepo    *repository.UserRepository
}

func NewAIService(
	apiKey string,
	matRepo *repository.MaterialRepository,
	pkgRepo *repository.PackageRepository,
	userRepo *repository.UserRepository,
) (*AIService, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY tidak ditemukan")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	return &AIService{
		geminiModel: model,
		matRepo:     matRepo,
		pkgRepo:     pkgRepo,
		userRepo:    userRepo,
	}, nil
}

// --- Helper Validasi Package ---
func (s *AIService) validatePackage(pkgIDStr string, userID uint) (*uint, error) {
	if pkgIDStr == "" {
		return nil, nil
	}
	pkgID, err := strconv.ParseUint(pkgIDStr, 10, 32)
	if err != nil {
		return nil, errors.New("package_id tidak valid")
	}
	_, err = s.pkgRepo.FindByID(uint(pkgID), userID)
	if err != nil {
		return nil, errors.New("package tidak ditemukan atau anda bukan pemiliknya")
	}
	finalID := uint(pkgID)
	return &finalID, nil
}

func (s *AIService) IngestPDF(ctx context.Context, fileHeader *multipart.FileHeader, title, packageIDStr, userIDString string) (*dto.MaterialResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	pkgID, err := s.validatePackage(packageIDStr, uint(userID))
	if err != nil {
		return nil, err
	}

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

	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         title,
		SourceType:    "pdf",
		Source:        fileHeader.Filename,
		ExtractedText: rawText,
		PackageID:     pkgID,
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
	userID, _ := strconv.ParseUint(userIDString, 10, 32)
	var pkgID *uint

	if req.PackageID != nil {
		_, err := s.pkgRepo.FindByID(*req.PackageID, uint(userID))
		if err != nil {
			return nil, errors.New("package tidak ditemukan atau anda bukan pemiliknya")
		}
		pkgID = req.PackageID
	}

	rawText, err := utils.ExtractTextFromYouTube(req.URL)
	if err != nil {
		return nil, err
	}
	if rawText == "" {
		return nil, errors.New("video ini tidak memiliki transkrip")
	}

	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         req.Title,
		SourceType:    "youtube",
		Source:        req.URL,
		ExtractedText: rawText,
		PackageID:     pkgID,
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

	if material.Summary != "" {
		log.Printf("Summary materi ID %d diambil dari database (Cache Hit)", material.ID)
		return &dto.GenerateSummaryResponse{
			MaterialID: material.ID,
			Summary:    material.Summary,
		}, nil
	}

	log.Printf("Summary materi ID %d belum ada. Meminta Gemini...", material.ID)

	prompt := fmt.Sprintf("Jelaskan ulang isi materi berikut secara jelas, mendalam, dan terstruktur, seperti seorang dosen profesional... Materi:\n\n%s", material.ExtractedText)

	s.geminiModel.ResponseMIMEType = "text/plain"
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

	material.Summary = summary
	if err := s.matRepo.Update(material); err != nil {
		log.Printf("Gagal menyimpan summary ke DB: %v", err)
	}

	return &dto.GenerateSummaryResponse{
		MaterialID: material.ID,
		Summary:    summary,
	}, nil
}

func (s *AIService) GenerateFlashcards(ctx context.Context, req dto.GenerateFlashcardRequest, userIDString string) (*dto.GenerateFlashcardResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak punya akses")
	}

	prompt := fmt.Sprintf(`
		Buatkan flashcard (kartu belajar) berdasarkan materi berikut.
		Flashcard terdiri dari "front" (pertanyaan ringkas atau istilah) dan "back" (jawaban atau definisi penjelasan).
		
		Format Output HARUS JSON ARRAY MURNI:
		[
			{"front": "Apa itu X?", "back": "X adalah..."},
			{"front": "Istilah Y", "back": "Definisi Y..."}
		]

		Instruksi:
		1. Gunakan bahasa Indonesia yang jelas.
		2. "front" harus singkat dan memancing ingatan.
		3. "back" harus padat dan menjelaskan inti konsep.
		4. Jangan tambahkan markdown code block.

		Tambahkan untuk jumlah di flash cardnya sesuai isi materi saja, jika materi memungkinkan membuat 10 flash card buatkan 10 kalau cocoknya buat 5 ya kasih 5 jadi sesuaikan agar semua yang ada di materi masuk ke flash card. 
		Materi:
		%s
	`, material.ExtractedText)

	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal memanggil Gemini: %w", err)
	}

	var jsonString string
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				jsonString += string(txt)
			}
		}
	}

	jsonString = strings.TrimSpace(jsonString)
	jsonString = strings.TrimPrefix(jsonString, "```json")
	jsonString = strings.TrimSuffix(jsonString, "```")

	var cards []dto.FlashcardItem
	if err := json.Unmarshal([]byte(jsonString), &cards); err != nil {
		log.Printf("Error parsing flashcards: %v | Raw: %s", err, jsonString)
		return nil, errors.New("gagal memproses hasil flashcard dari AI")
	}

	return &dto.GenerateFlashcardResponse{
		MaterialID: req.MaterialID,
		Flashcards: cards,
	}, nil
}

func (s *AIService) GenerateQuiz(ctx context.Context, req dto.GenerateQuizRequest, userIDString string) (*dto.GenerateQuizResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak punya akses")
	}

	prompt := fmt.Sprintf(`
		Buatkan %d soal latihan berdasarkan materi berikut.
		Output harus JSON Array.
		Materi: %s`, req.QuestionCount, material.ExtractedText)

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

	var questions []dto.QuizQuestion
	if err := json.Unmarshal([]byte(quizJSON), &questions); err != nil {
		return nil, fmt.Errorf("gagal parsing hasil Gemini: %w", err)
	}

	return &dto.GenerateQuizResponse{
		MaterialID: req.MaterialID,
		Questions:  questions,
	}, nil
}

func (s *AIService) GetDailyQuiz(ctx context.Context, userIDString string) (*dto.DailyQuizResponse, error) {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if user.LastStreakAwardedDate != nil {
		lastAward := *user.LastStreakAwardedDate
		lastAwardDate := time.Date(lastAward.Year(), lastAward.Month(), lastAward.Day(), 0, 0, 0, 0, time.Local)

		if lastAwardDate.Equal(today) {
			return &dto.DailyQuizResponse{
				Questions: nil,
				IsDone:    true,
			}, nil
		}
	}

	count, err := s.matRepo.CountByUserID(uint(userID))
	if count == 0 {
		return nil, errors.New("anda belum mengupload materi apapun")
	}

	materials, err := s.matRepo.FindRandomByUserID(uint(userID), 3)
	if err != nil {
		return nil, errors.New("gagal mengambil materi acak")
	}

	var combinedText strings.Builder
	for _, m := range materials {
		combinedText.WriteString(m.ExtractedText)
		combinedText.WriteString("\n\n")
	}
	finalText := combinedText.String()
	if len(finalText) > 12000 {
		finalText = finalText[:12000]
	}

	prompt := fmt.Sprintf(`
		Buatkan 10 soal pilihan ganda berdasarkan teks gabungan ini.
		
		Format Output HARUS JSON ARRAY (Strict JSON):
		[
			{
				"id": 1,
				"pertanyaan": "...",
				"pilihan": [
					{"A": "opsi A"}, {"B": "opsi B"}, {"C": "opsi C"}, {"D": "opsi D"}
				],
				"jawaban_benar": "A"
			}
		]

		Pastikan JSON valid tanpa markdown tambahan.
		Materi: %s
	`, finalText)

	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi AI: %w", err)
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

	var checkJSON []map[string]interface{}
	if err := json.Unmarshal([]byte(quizJSON), &checkJSON); err != nil {
		return nil, errors.New("AI menghasilkan format soal yang tidak valid, silakan coba lagi")
	}

	return &dto.DailyQuizResponse{
		Questions: checkJSON,
		IsDone:    false,
	}, nil
}

func (s *AIService) ClaimDailyStreak(userIDString string) error {
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return errors.New("user ID tidak valid")
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if user.LastStreakAwardedDate != nil {
		lastAward := *user.LastStreakAwardedDate
		lastAwardDate := time.Date(lastAward.Year(), lastAward.Month(), lastAward.Day(), 0, 0, 0, 0, time.Local)

		if lastAwardDate.Equal(today) {
			return errors.New("anda sudah mengklaim streak hari ini")
		}
	}

	user.CurrentStreak += 1
	user.LastStreakAwardedDate = &now

	if _, err := s.userRepo.Update(user); err != nil {
		return errors.New("gagal menambahkan poin streak")
	}

	log.Printf("STREAK UP! User %d - Total Streak: %d", userID, user.CurrentStreak)
	return nil
}
