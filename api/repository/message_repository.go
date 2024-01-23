package repository

import (
	"zapmeow/api/model"
	"zapmeow/pkg/database"
)

type MessageRepository interface {
	CreateMessage(message *model.Message) error
	CreateMessages(messages *[]model.Message) error
	GetChatMessages(instanceID string, chatJID string) (*[]model.Message, error)
	CountChatMessages(instanceID string, chatJID string) (int64, error)
	DeleteMessagesByInstanceID(instanceID string) error
}

type messageRepository struct {
	database database.Database
}

func NewMessageRepository(database database.Database) *messageRepository {
	return &messageRepository{database: database}
}

func (repo *messageRepository) CreateMessage(message *model.Message) error {
	return repo.database.Client().Create(message).Error
}

func (repo *messageRepository) CreateMessages(messages *[]model.Message) error {
	return repo.database.Client().Create(messages).Error
}

func (repo *messageRepository) CountChatMessages(instanceID string, chatJID string) (int64, error) {
	var count int64
	if result := repo.database.Client().Model(&model.Message{}).Where("instance_id = ? AND chat_jid = ?", instanceID, chatJID).Count(&count); result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (repo *messageRepository) GetChatMessages(instanceID string, chatJID string) (*[]model.Message, error) {
	var messages []model.Message
	if result := repo.database.Client().Where("instance_id = ? AND chat_jid = ?", instanceID, chatJID).Order("timestamp DESC").Find(&messages); result.Error != nil {
		return nil, result.Error
	}
	return &messages, nil
}

func (repo *messageRepository) DeleteMessagesByInstanceID(instanceID string) error {
	if result := repo.database.Client().Where("instance_id = ?", instanceID).Unscoped().Delete(&model.Message{}); result.Error != nil {
		return result.Error
	}
	return nil
}
