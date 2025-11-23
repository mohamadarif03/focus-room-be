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
	"gorm.io/datatypes"
)

type AIService struct {
	genaiClient *genai.Client
	geminiModel *genai.GenerativeModel
	matRepo     *repository.MaterialRepository
	pkgRepo     *repository.PackageRepository
	userRepo    *repository.UserRepository
	statsRepo   *repository.StatsRepository
	quizRepo    *repository.QuizRepository
}

func NewAIService(
	apiKey string,
	matRepo *repository.MaterialRepository,
	pkgRepo *repository.PackageRepository,
	userRepo *repository.UserRepository,
	statsRepo *repository.StatsRepository,
	quizRepo *repository.QuizRepository,
) (*AIService, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY tidak ditemukan")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	// Model Utama (Default JSON untuk Quiz/Flashcard)
	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	return &AIService{
		genaiClient: client,
		geminiModel: model,
		matRepo:     matRepo,
		pkgRepo:     pkgRepo,
		userRepo:    userRepo,
		statsRepo:   statsRepo,
		quizRepo:    quizRepo,
	}, nil
}

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

// --- Helper Prompt Dosen (Agar Konsisten) ---
func getLecturerPrompt(text string) string {
	return fmt.Sprintf(`
Jelaskan ulang isi materi berikut secara jelas, mendalam, dan terstruktur, seperti seorang dosen profesional yang menjelaskan konsep di kelas, namun tanpa sapaan pembuka atau penutup.

Saat menjelaskan ulang:
1. Gunakan bahasa yang natural, komunikatif, dan logis.
2. Fokus untuk memperjelas isi materi, bukan sekadar merangkum.
3. Jelaskan konsep dan ide utama dengan contoh nyata atau analogi jika perlu.
4. Tutup dengan ringkasan inti dan kesimpulan.

Materi:
%s`, text)
}

func (s *AIService) GetMaterials(userIDString, userRole, packageIDString string) ([]dto.MaterialResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	var pkgID *uint
	if packageIDString != "" {
		id, err := strconv.ParseUint(packageIDString, 10, 32)
		if err == nil {
			uid := uint(id)
			pkgID = &uid
		}
	}

	materials, err := s.matRepo.FindAll(uint(userID), userRole, pkgID)
	if err != nil {
		return nil, err
	}

	var responses []dto.MaterialResponse
	for _, m := range materials {
		pkgTitle := ""
		if m.PackageID != nil {
			pkgTitle = m.Package.Title
		}
		responses = append(responses, dto.MaterialResponse{
			ID:           m.ID,
			Title:        m.Title,
			SourceType:   m.SourceType,
			Source:       m.Source,
			Summary:      m.Summary,
			IsPublic:     m.IsPublic,
			PackageTitle: pkgTitle,
		})
	}

	return responses, nil
}

func (s *AIService) IngestPDF(ctx context.Context, fileHeader *multipart.FileHeader, packageIDStr, userIDString, userRole string) (*dto.MaterialResponse, error) {
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

	finalTitle := fileHeader.Filename

	rawText, err := utils.ExtractTextFromPDF(file, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("gagal ekstrak PDF: %w", err)
	}
	if rawText == "" {
		return nil, errors.New("PDF ini tidak mengandung teks")
	}

	summaryModel := s.genaiClient.GenerativeModel("gemini-2.5-flash")
	summaryModel.ResponseMIMEType = "text/plain"

	inputText := rawText
	if len(inputText) > 30000 {
		inputText = inputText[:30000] + "..."
	}

	prompt := getLecturerPrompt(inputText)
	summary := ""

	resp, err := summaryModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Warning: Gagal auto-summary PDF: %v", err)
	} else {
		if len(resp.Candidates) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					summary += string(txt)
				}
			}
		}
	}

	isPublic := false
	if userRole == "admin" {
		isPublic = true
	}

	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         finalTitle,
		SourceType:    "pdf",
		Source:        fileHeader.Filename,
		ExtractedText: rawText,
		Summary:       summary,
		PackageID:     pkgID,
		IsPublic:      isPublic,
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
		IsPublic:   savedMat.IsPublic,
	}, nil
}

func (s *AIService) GetMaterialDetail(idString, userIDString, userRole string) (*dto.MaterialResponse, error) {
	id, err := strconv.ParseUint(idString, 10, 32)
	if err != nil {
		return nil, errors.New("ID materi tidak valid")
	}
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindOneAccessible(uint(id), uint(userID), userRole)
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak memiliki akses")
	}
	pkgTitle := ""
	if material.PackageID != nil {
		pkgTitle = material.Package.Title
	}

	return &dto.MaterialResponse{
		ID:           material.ID,
		Title:        material.Title,
		SourceType:   material.SourceType,
		Source:       material.Source,
		Summary:      material.Summary,
		IsPublic:     material.IsPublic,
		PackageTitle: pkgTitle,
		CreatedAt:    material.CreatedAt,
	}, nil
}

func (s *AIService) IngestYouTube(ctx context.Context, req dto.IngestYouTubeRequest, userIDString, userRole string) (*dto.MaterialResponse, error) {
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

	// Ambil Transkrip
	rawText, err := utils.ExtractTextFromYouTube(req.URL)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil transkrip: %w (pastikan video punya CC/Subtitle)", err)
	}
	if rawText == "" {
		return nil, errors.New("video ini tidak memiliki teks transkrip")
	}

	summaryModel := s.genaiClient.GenerativeModel("gemini-2.5-flash")
	summaryModel.ResponseMIMEType = "text/plain"

	inputText := rawText
	if len(inputText) > 30000 {
		inputText = inputText[:30000] + "..."
	}

	prompt := getLecturerPrompt(inputText)
	summary := ""

	resp, err := summaryModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Warning: Gagal auto-summary YouTube: %v", err)
	} else {
		if len(resp.Candidates) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					summary += string(txt)
				}
			}
		}
	}

	isPublic := false
	if userRole == "admin" {
		isPublic = true
	}

	newMaterial := &model.Material{
		UserID:        uint(userID),
		Title:         title,
		SourceType:    "youtube",
		Source:        req.URL,
		ExtractedText: rawText,
		Summary:       summary,
		PackageID:     pkgID,
		IsPublic:      isPublic,
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
		IsPublic:   savedMat.IsPublic,
	}, nil
}

func (s *AIService) GenerateSummary(ctx context.Context, req dto.GenerateSummaryRequest, userIDString string) (*dto.GenerateSummaryResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau anda tidak punya akses")
	}

	if material.Summary != "" {
		return &dto.GenerateSummaryResponse{
			MaterialID: material.ID,
			Summary:    material.Summary,
		}, nil
	}

	summaryModel := s.genaiClient.GenerativeModel("gemini-2.5-flash")
	summaryModel.ResponseMIMEType = "text/plain"

	prompt := getLecturerPrompt(material.ExtractedText)

	resp, err := summaryModel.GenerateContent(ctx, genai.Text(prompt))
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
	s.matRepo.Update(material)

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
		Format Output HARUS JSON ARRAY MURNI:
		[{"front": "...", "back": "..."}]
		Jumlah kartu: Sesuai kecocokan materi (max 10).
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
		Buatkan %d soal pilihan ganda berdasarkan materi berikut.
		
		Format Output HARUS JSON ARRAY MURNI dengan struktur pilihan Key-Value:
		[
			{
				"pertanyaan": "...", 
				"pilihan": [
					{"key": "A", "value": "Jawaban A"},
					{"key": "B", "value": "Jawaban B"},
					{"key": "C", "value": "Jawaban C"},
					{"key": "D", "value": "Jawaban D"}
				],
				"jawaban_benar": "A"
			}
		]

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

	var rawQuestions []dto.QuizQuestion
	if err := json.Unmarshal([]byte(quizJSON), &rawQuestions); err != nil {
		return nil, fmt.Errorf("gagal parsing hasil Gemini: %w", err)
	}

	var dbQuestions []model.QuizQuestion
	for _, q := range rawQuestions {
		pilihanJSON, err := json.Marshal(q.Pilihan)
		if err != nil {
			continue
		}

		dbQuestions = append(dbQuestions, model.QuizQuestion{
			Pertanyaan:   q.Pertanyaan,
			Pilihan:      datatypes.JSON(pilihanJSON),
			JawabanBenar: q.JawabanBenar,
		})
	}

	newQuiz := &model.Quiz{
		MaterialID: material.ID,
		Questions:  dbQuestions,
	}

	if err := s.quizRepo.CreateWithLimit(newQuiz); err != nil {
		return nil, fmt.Errorf("gagal menyimpan quiz ke database: %w", err)
	}

	var responseQuestions []dto.QuizQuestion
	for _, q := range newQuiz.Questions {

		var pilihanItems []dto.OptionItem
		json.Unmarshal(q.Pilihan, &pilihanItems)

		responseQuestions = append(responseQuestions, dto.QuizQuestion{
			ID:           int(q.ID),
			Pertanyaan:   q.Pertanyaan,
			Pilihan:      pilihanItems,
			JawabanBenar: q.JawabanBenar,
		})
	}

	return &dto.GenerateQuizResponse{
		ID:         int(newQuiz.ID),
		MaterialID: material.ID,
		Questions:  responseQuestions,
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

	// Gunakan s.geminiModel (Default: JSON)
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

// --- CLAIM STREAK ---
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

	quizLog := &model.QuizLog{
		UserID:    user.ID,
		Score:     10,
		CreatedAt: time.Now(),
	}
	if err := s.statsRepo.CreateQuizLog(quizLog); err != nil {
		log.Printf("Warning: Gagal log quiz activity: %v", err)
	}

	log.Printf("STREAK UP! User %d - Total Streak: %d", userID, user.CurrentStreak)
	return nil
}

func (s *AIService) GetDailyQuizStatus(userIDString string) (*dto.DailyQuizStatusResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	isDone := false
	if user.LastStreakAwardedDate != nil {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

		lastAward := *user.LastStreakAwardedDate
		lastAwardDate := time.Date(lastAward.Year(), lastAward.Month(), lastAward.Day(), 0, 0, 0, 0, time.Local)

		if lastAwardDate.Equal(today) {
			isDone = true
		}
	}

	return &dto.DailyQuizStatusResponse{
		IsDone: isDone,
		Streak: user.CurrentStreak,
	}, nil
}
