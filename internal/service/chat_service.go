package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv" // Pastikan ada import ini
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/mohamadarif03/focus-room-be/internal/dto"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"google.golang.org/api/option"
)

type ChatService struct {
	chatRepo    *repository.ChatRepository
	matRepo     *repository.MaterialRepository
	genaiClient *genai.Client
}

func NewChatService(apiKey string, chatRepo *repository.ChatRepository, matRepo *repository.MaterialRepository) (*ChatService, error) {
	if apiKey == "" {
		return nil, errors.New("API Key kosong")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &ChatService{
		chatRepo:    chatRepo,
		matRepo:     matRepo,
		genaiClient: client,
	}, nil
}

func (s *ChatService) SendChat(ctx context.Context, req dto.ChatRequest, userIDString string) (*dto.ChatResponse, error) {
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	material, err := s.matRepo.FindByID(req.MaterialID, uint(userID))
	if err != nil {
		return nil, errors.New("materi tidak ditemukan atau akses ditolak")
	}

	userChat := &model.MaterialChat{
		MaterialID: req.MaterialID,
		UserID:     uint(userID),
		Role:       "user",
		Message:    req.Message,
		CreatedAt:  time.Now(),
	}
	if err := s.chatRepo.Create(userChat); err != nil {
		return nil, errors.New("gagal menyimpan pesan")
	}

	genModel := s.genaiClient.GenerativeModel("gemini-2.5-flash")
	genModel.ResponseMIMEType = "text/plain"

	prompt := fmt.Sprintf(`
		Kamu adalah asisten belajar. Jawab pertanyaan user berdasarkan materi berikut.
		Jika jawaban tidak ada di materi, katakan "Maaf, tidak ada info tersebut di materi ini."
		
		=== MATERI ===
		%s
		=== END MATERI ===

		User: %s
	`, material.ExtractedText, req.Message)

	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))

	var aiReply string

	if err != nil {
		aiReply = "Maaf, AI sedang sibuk atau terjadi gangguan koneksi."
		log.Printf("Gemini Error: %v", err)
	} else if len(resp.Candidates) == 0 {
		aiReply = "Maaf, AI tidak memberikan respon."
	} else {
		// Jika sukses, ambil isinya
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				aiReply += string(txt)
			}
		}
	}

	aiChat := &model.MaterialChat{
		MaterialID: req.MaterialID,
		UserID:     uint(userID),
		Role:       "assistant",
		Message:    aiReply,
		CreatedAt:  time.Now(),
	}
	s.chatRepo.Create(aiChat)

	return &dto.ChatResponse{
		ID:        aiChat.ID,
		Reply:     aiChat.Message,
		CreatedAt: aiChat.CreatedAt,
	}, nil
}

func (s *ChatService) GetHistory(materialIDString, userIDString string) ([]dto.ChatHistoryItem, error) {
	materialID, _ := strconv.ParseUint(materialIDString, 10, 32)
	userID, _ := strconv.ParseUint(userIDString, 10, 32)

	if _, err := s.matRepo.FindByID(uint(materialID), uint(userID)); err != nil {
		return nil, errors.New("akses ditolak")
	}

	chats, err := s.chatRepo.FindByMaterialAndUser(uint(materialID), uint(userID))
	if err != nil {
		return nil, err
	}

	var response []dto.ChatHistoryItem
	for _, c := range chats {
		response = append(response, dto.ChatHistoryItem{
			ID:        c.ID,
			Role:      c.Role,
			Message:   c.Message,
			CreatedAt: c.CreatedAt,
		})
	}
	return response, nil
}
