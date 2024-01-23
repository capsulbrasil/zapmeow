package service

import (
	"zapmeow/api/model"
	"zapmeow/api/repository"
)

type MessageService interface {
	CreateMessage(message *model.Message) error
	CreateMessages(messages *[]model.Message) error
	GetChatMessages(instanceID string, chatJID string) (*[]model.Message, error)
	CountChatMessages(instanceID string, chatJID string) (int64, error)
	DeleteMessagesByInstanceID(instanceID string) error
}

type messageService struct {
	messageRep repository.MessageRepository
}

func NewMessageService(messageRep repository.MessageRepository) *messageService {
	return &messageService{
		messageRep: messageRep,
	}
}

func (m *messageService) CreateMessage(message *model.Message) error {
	return m.messageRep.CreateMessage(message)
}

func (m *messageService) CreateMessages(messages *[]model.Message) error {
	return m.messageRep.CreateMessages(messages)
}

func (m *messageService) GetChatMessages(instanceID string, chatJID string) (*[]model.Message, error) {
	return m.messageRep.GetChatMessages(instanceID, chatJID)
}

func (m *messageService) CountChatMessages(instanceID string, chatJID string) (int64, error) {
	return m.messageRep.CountChatMessages(instanceID, chatJID)
}

func (m *messageService) DeleteMessagesByInstanceID(instanceID string) error {
	return m.messageRep.DeleteMessagesByInstanceID(instanceID)
}
