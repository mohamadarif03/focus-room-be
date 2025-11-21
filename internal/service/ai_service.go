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

	// Default MIME Type kita set ke JSON (untuk fitur Quiz/Flashcard)
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

func getLecturerPrompt(text string) string {
	return fmt.Sprintf(`
Jelaskan ulang isi materi berikut secara jelas, mendalam, dan terstruktur, seperti seorang dosen profesional yang menjelaskan konsep di kelas, namun tanpa sapaan pembuka atau penutup kelas (misalnya: “Selamat pagi mahasiswa”, “Apakah ada pertanyaan?”, dan sejenisnya).

Saat menjelaskan ulang:
1. Gunakan bahasa yang natural, komunikatif, dan logis, bukan formal kaku.
2. Fokus untuk memperjelas isi materi, bukan sekadar merangkum.
3. Jelaskan konsep dan ide utama dengan contoh nyata atau analogi jika perlu.
4. Jika ada istilah sulit, jelaskan maknanya terlebih dahulu sebelum lanjut.
5. Gunakan gaya penjelasan yang mengalir seperti narasi dosen yang fokus menjelaskan isi (tanpa salam, tanpa tanya jawab).
6. Tutup dengan ringkasan inti dan kesimpulan.

Output yang diharapkan:
- Penjelasan ulang yang runtut, detail, dan mudah dipahami.
- Gaya profesional namun tetap natural.
- Tidak ada bagian sapaan, humor, atau tanya-jawab interaktif.

Materi:
%s`, text)
}

func (s *AIService) IngestPDF(ctx context.Context, fileHeader *multipart.FileHeader, title, packageIDStr, userIDString string) (*dto.MaterialResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	// 1. Validasi Package
	pkgID, err := s.validatePackage(packageIDStr, uint(userID))
	if err != nil {
		return nil, err
	}

	// 2. Buka File
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("gagal membuka file")
	}
	defer file.Close()

	// 3. Ekstrak PDF
	rawText, err := utils.ExtractTextFromPDF(file, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("gagal baca PDF: %w", err)
	}

	// 4. Fallback Title: Jika title kosong, pakai nama file
	finalTitle := title
	if finalTitle == "" {
		finalTitle = fileHeader.Filename
	}

	// 5. Generate Summary
	// PENTING: Clone model atau set konfigurasi ulang agar tidak bentrok setting JSON/Text
	s.geminiModel.ResponseMIMEType = "text/plain"

	// Potong text jika terlalu panjang agar tidak error token limit
	inputText := rawText
	if len(inputText) > 30000 {
		inputText = inputText[:30000] + "..."
	}

	prompt := getLecturerPrompt(inputText)
	summary := ""

	// Generate Content
	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))

	// CEK ERROR: Jangan sekadar di-log jika ingin memastikan user tahu
	if err != nil {
		log.Printf("ERROR GEMINI: %v", err)
		// Opsi A: Return error (User gagal upload)
		// return nil, fmt.Errorf("gagal generate summary: %v", err)

		// Opsi B: Lanjut tapi summary kosong (User berhasil upload, summary nanti)
		summary = "Gagal membuat rangkuman otomatis. Silakan coba generate ulang."
	} else {
		if len(resp.Candidates) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					summary += string(txt)
				}
			}
		}
	}

	// 6. Simpan ke DB
	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         finalTitle, // Pakai finalTitle
		SourceType:    "pdf",
		Source:        fileHeader.Filename,
		ExtractedText: rawText,
		Summary:       summary,
		PackageID:     pkgID,
	}

	savedMat, err := s.matRepo.Save(newMaterial)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan materi: %w", err)
	}

	return &dto.MaterialResponse{
		ID:         savedMat.ID,
		Title:      savedMat.Title,
		SourceType: savedMat.SourceType,
		Source:     savedMat.Source,
		Summary:    savedMat.Summary,
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

	title := req.Title
	if title == "" {
		fetchedTitle, err := utils.GetVideoTitle(req.URL)
		if err != nil {
			title = req.URL
		} else {
			title = fetchedTitle
		}
	}

	rawText, err := utils.ExtractTextFromYouTube(req.URL)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil transkrip: %w (pastikan video punya CC/Subtitle)", err)
	}
	if rawText == "" {
		return nil, errors.New("video ini tidak memiliki teks transkrip")
	}

	s.geminiModel.ResponseMIMEType = "text/plain"

	inputText := rawText
	if len(inputText) > 30000 {
		inputText = inputText[:30000] + "..."
	}

	prompt := getLecturerPrompt(inputText)
	summary := ""

	resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Warning: Gagal auto-summary YouTube: %v", err)
		summary = "Gagal membuat rangkuman otomatis. Anda bisa mencoba generate ulang nanti."
	} else {
		if len(resp.Candidates) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					summary += string(txt)
				}
			}
		}
	}

	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         title,
		SourceType:    "youtube",
		Source:        req.URL,
		ExtractedText: rawText,
		Summary:       summary,
		PackageID:     pkgID,
	}

	savedMat, err := s.matRepo.Save(newMaterial)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan materi: %w", err)
	}

	return &dto.MaterialResponse{
		ID:         savedMat.ID,
		Title:      savedMat.Title,
		SourceType: savedMat.SourceType,
		Source:     savedMat.Source,
		Summary:    savedMat.Summary,
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

	s.geminiModel.ResponseMIMEType = "text/plain"

	prompt := getLecturerPrompt(material.ExtractedText)

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

	s.geminiModel.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`
		Buatkan flashcard (kartu belajar) berdasarkan materi berikut.
		Format Output HARUS JSON ARRAY MURNI:
		[{"front": "...", "back": "..."}]
		Materi: %s
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

	s.geminiModel.ResponseMIMEType = "application/json"

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

	// Ubah mode ke JSON
	s.geminiModel.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`
		Buatkan 10 soal pilihan ganda berdasarkan teks gabungan ini.
		Format Output HARUS JSON ARRAY (Strict JSON):
		[{"id": 1, "pertanyaan": "...", "pilihan": [{"A": ".."}], "jawaban_benar": "A"}]
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

func (s *AIService) GetMaterials(userIDString, packageIDString string) ([]dto.MaterialResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	var pkgID *uint
	if packageIDString != "" {
		id, err := strconv.ParseUint(packageIDString, 10, 32)
		if err == nil {
			uid := uint(id)
			pkgID = &uid
		}
	}

	materials, err := s.matRepo.FindAll(uint(userID), pkgID)
	if err != nil {
		return nil, err
	}

	var responses []dto.MaterialResponse
	for _, m := range materials {
		responses = append(responses, dto.MaterialResponse{
			ID:         m.ID,
			Title:      m.Title,
			SourceType: m.SourceType,
			Source:     m.Source,
			Summary:    m.Summary,
		})
	}

	return responses, nil
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

	// quizLog := &model.QuizLog{
	// 	UserID:    user.ID,
	// 	Score:     10,
	// 	CreatedAt: time.Now(),
	// }
	// _ = quizLog

	log.Printf("STREAK UP! User %d - Total Streak: %d", userID, user.CurrentStreak)
	return nil
}
