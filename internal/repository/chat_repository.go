package repository

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(chat *model.MaterialChat) error {
	return r.db.Create(chat).Error
}

func (r *ChatRepository) FindByMaterialAndUser(materialID, userID uint) ([]model.MaterialChat, error) {
	var chats []model.MaterialChat
	err := r.db.Where("material_id = ? AND user_id = ?", materialID, userID).
		Order("created_at asc").
		Find(&chats).Error
	return chats, err
}
