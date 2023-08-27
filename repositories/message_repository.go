package repositories

import (
	"zapmeow/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	CreateMessage(message *models.Message) error
	CreateMessages(message *[]models.Message) error
	CountMessages(chatJID string, meJID string) (int64, error)
	GetMessages(chatJID string, meJID string) (*[]models.Message, error)
	DeleteMessagesByChatJID(chatId string) error
}

type DatabaseMessageRepository struct {
	db *gorm.DB
}

func NewDatabaseMessageRepository(db *gorm.DB) *DatabaseMessageRepository {
	return &DatabaseMessageRepository{db: db}
}

func (repo *DatabaseMessageRepository) CreateMessage(message *models.Message) error {
	return repo.db.Create(message).Error
}

func (repo *DatabaseMessageRepository) CreateMessages(messages *[]models.Message) error {
	return repo.db.Create(messages).Error
}

func (repo *DatabaseMessageRepository) CountMessages(chatJID string, meJID string) (int64, error) {
	var count int64
	if result := repo.db.Model(&models.Message{}).Where("chat_jid = ? AND me_jid = ?", chatJID, meJID).Count(&count); result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (repo *DatabaseMessageRepository) GetMessages(chatJID string, meJID string) (*[]models.Message, error) {
	var messages []models.Message
	if result := repo.db.Where("chat_jid = ? AND me_jid = ?", chatJID, meJID).Order("timestamp DESC").Find(&messages); result.Error != nil {
		return nil, result.Error
	}
	return &messages, nil
}

func (repo *DatabaseMessageRepository) DeleteMessagesByChatJID(chatJID string) error {
	if result := repo.db.Where("chat_jid = ?", chatJID).Delete(&models.Message{}); result.Error != nil {
		return result.Error
	}
	return nil
}
