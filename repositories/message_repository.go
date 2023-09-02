package repositories

import (
	"zapmeow/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	CreateMessage(message *models.Message) error
	CreateMessages(messages *[]models.Message) error
	GetChatMessages(chatJID string, meJID string) (*[]models.Message, error)
	CountChatMessages(chatJID string, meJID string) (int64, error)
	DeleteMessagesByChatJID(chatJID string) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *messageRepository {
	return &messageRepository{db: db}
}

func (repo *messageRepository) CreateMessage(message *models.Message) error {
	return repo.db.Create(message).Error
}

func (repo *messageRepository) CreateMessages(messages *[]models.Message) error {
	return repo.db.Create(messages).Error
}

func (repo *messageRepository) CountChatMessages(chatJID string, meJID string) (int64, error) {
	var count int64
	if result := repo.db.Model(&models.Message{}).Where("chat_jid = ? AND me_jid = ?", chatJID, meJID).Count(&count); result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (repo *messageRepository) GetChatMessages(chatJID string, meJID string) (*[]models.Message, error) {
	var messages []models.Message
	if result := repo.db.Where("chat_jid = ? AND me_jid = ?", chatJID, meJID).Order("timestamp DESC").Find(&messages); result.Error != nil {
		return nil, result.Error
	}
	return &messages, nil
}

func (repo *messageRepository) DeleteMessagesByChatJID(chatJID string) error {
	if result := repo.db.Where("chat_jid = ?", chatJID).Delete(&models.Message{}); result.Error != nil {
		return result.Error
	}
	return nil
}
