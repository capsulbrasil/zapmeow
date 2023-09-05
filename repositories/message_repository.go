package repositories

import (
	"zapmeow/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	CreateMessage(message *models.Message) error
	CreateMessages(messages *[]models.Message) error
	GetChatMessages(instanceID string, chatJID string) (*[]models.Message, error)
	CountChatMessages(instanceID string, chatJID string) (int64, error)
	DeleteMessagesByInstanceID(instanceID string) error
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

func (repo *messageRepository) CountChatMessages(instanceID string, chatJID string) (int64, error) {
	var count int64
	if result := repo.db.Model(&models.Message{}).Where("instance_id = ? AND chat_jid = ?", instanceID, chatJID).Count(&count); result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (repo *messageRepository) GetChatMessages(instanceID string, chatJID string) (*[]models.Message, error) {
	var messages []models.Message
	if result := repo.db.Where("instance_id = ? AND chat_jid = ?", instanceID, chatJID).Order("timestamp DESC").Find(&messages); result.Error != nil {
		return nil, result.Error
	}
	return &messages, nil
}

func (repo *messageRepository) DeleteMessagesByInstanceID(instanceID string) error {
	if result := repo.db.Where("instance_id = ?", instanceID).Unscoped().Delete(&models.Message{}); result.Error != nil {
		return result.Error
	}
	return nil
}
