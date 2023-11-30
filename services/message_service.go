package services

import (
	"encoding/base64"
	"io/ioutil"
	"mime"
	"path/filepath"
	"time"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/repositories"
)

type MessageService interface {
	CreateMessage(message *models.Message) error
	CreateMessages(messages *[]models.Message) error
	GetChatMessages(instanceID string, chatJID string) (*[]models.Message, error)
	CountChatMessages(instanceID string, chatJID string) (int64, error)
	DeleteMessagesByInstanceID(instanceID string) error
	ToJSON(message models.Message) Message
}

type Message struct {
	ID        uint
	Sender    string
	Chat      string
	MessageID string
	FromMe    bool
	Timestamp time.Time
	Body      string
	MediaType string
	MediaData *struct {
		Mimetype string
		Base64   string
	}
}

type messageService struct {
	messageRep repositories.MessageRepository
	log        configs.Logger
}

func NewMessageService(messageRep repositories.MessageRepository, log configs.Logger) *messageService {
	return &messageService{
		messageRep: messageRep,
		log:        log,
	}
}

func (m *messageService) CreateMessage(message *models.Message) error {
	return m.messageRep.CreateMessage(message)
}

func (m *messageService) CreateMessages(messages *[]models.Message) error {
	return m.messageRep.CreateMessages(messages)
}

func (m *messageService) GetChatMessages(instanceID string, chatJID string) (*[]models.Message, error) {
	return m.messageRep.GetChatMessages(instanceID, chatJID)
}

func (m *messageService) CountChatMessages(instanceID string, chatJID string) (int64, error) {
	return m.messageRep.CountChatMessages(instanceID, chatJID)
}

func (m *messageService) DeleteMessagesByInstanceID(instanceID string) error {
	return m.messageRep.DeleteMessagesByInstanceID(instanceID)
}

func (m *messageService) ToJSON(message models.Message) Message {
	messageJson := Message{
		ID:        message.ID,
		Sender:    message.SenderJID,
		Chat:      message.ChatJID,
		MessageID: message.MessageID,
		FromMe:    message.FromMe,
		Timestamp: message.Timestamp,
		Body:      message.Body,
		MediaType: message.MediaType,
	}

	if message.MediaType != "" {
		data, err := ioutil.ReadFile(message.MediaPath)
		if err != nil {
			m.log.Error("Error reading the file. ", err)
		} else {
			mimetype := mime.TypeByExtension(filepath.Ext(message.MediaPath))
			base64 := base64.StdEncoding.EncodeToString(data)
			messageJson.MediaData = &struct {
				Mimetype string
				Base64   string
			}{
				Mimetype: mimetype,
				Base64:   base64,
			}
		}
	}

	return messageJson
}
