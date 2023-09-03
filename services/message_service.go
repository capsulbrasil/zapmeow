package services

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/repositories"
	"zapmeow/utils"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
)

type MessageService interface {
	CreateMessage(message *models.Message) error
	CreateMessages(messages *[]models.Message) error
	GetChatMessages(instanceID string, chatJID string) (*[]models.Message, error)
	CountChatMessages(instanceID string, chatJID string) (int64, error)
	DeleteMessagesByInstanceID(instanceID string) error
	Parse(instance *configs.Instance, msg *events.Message) *models.Message
	ToJSON(message models.Message) map[string]interface{}
}

type messageService struct {
	messageRep repositories.MessageRepository
}

func NewMessageService(messageRep repositories.MessageRepository) *messageService {
	return &messageService{
		messageRep: messageRep,
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

func (m *messageService) Parse(instance *configs.Instance, msg *events.Message) *models.Message {
	mediaType, path := m.downloadMessageMedia(
		instance,
		msg.Message,
		msg.Info.ID,
	)

	var body = m.getTextMessage(msg.Message)
	if mediaType == "" && body == "" {
		return nil
	}

	if mediaType != "" {
		return &models.Message{
			InstanceID: instance.ID,
			MessageID:  msg.Info.ID,
			FromMe:     msg.Info.MessageSource.IsFromMe,
			ChatJID:    msg.Info.Chat.User,
			SenderJID:  msg.Info.Sender.User,
			Body:       body,
			MediaPath:  mediaType,
			MediaType:  path,
			Timestamp:  msg.Info.Timestamp,
		}
	}

	return &models.Message{
		InstanceID: instance.ID,
		MessageID:  msg.Info.ID,
		FromMe:     msg.Info.MessageSource.IsFromMe,
		ChatJID:    msg.Info.Chat.User,
		SenderJID:  msg.Info.Sender.User,
		Body:       body,
		Timestamp:  msg.Info.Timestamp,
	}
}

func (m *messageService) ToJSON(message models.Message) map[string]interface{} {
	messageJson := map[string]interface{}{
		"ID":        message.ID,
		"Sender":    message.SenderJID,
		"Chat":      message.ChatJID,
		"MessageID": message.MessageID,
		"FromMe":    message.FromMe,
		"Timestamp": message.Timestamp,
		"Body":      message.Body,
		"MediaType": message.MediaType,
	}

	if message.MediaType != "" {
		data, err := ioutil.ReadFile(message.MediaPath)
		if err != nil {
			fmt.Println(err)
		} else {
			mimetype := mime.TypeByExtension(filepath.Ext(message.MediaPath))
			base64 := base64.StdEncoding.EncodeToString(data)
			messageJson["MediaData"] = map[string]interface{}{
				"Mimetype": mimetype,
				"Base64":   base64,
			}
		}
	}

	return messageJson
}

func (m *messageService) downloadMidia() {

}

func (m *messageService) downloadMessageMedia(
	instance *configs.Instance,
	message *waProto.Message,
	fileName string,
) (string, string) {
	path := ""
	mediaType := ""

	dirPath := utils.MakeAccountStoragePath(instance.ID)
	err := os.MkdirAll(dirPath, 0751)
	if err != nil {
		return "", ""
	}

	document := message.GetDocumentMessage()
	if document != nil {
		mediaType = "document"

		data, err := instance.Client.Download(document)

		if err != nil {
			fmt.Println("Failed to download document", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(
			instance.ID,
			data,
			fileName,
			document.GetMimetype(),
		)
		if err != nil {
			fmt.Println("Failed to save document", err)
			return mediaType, ""
		}
		fmt.Println("Document saved")
	}

	audio := message.GetAudioMessage()
	if audio != nil {
		mediaType = "audio"

		data, err := instance.Client.Download(audio)
		if err != nil {
			fmt.Println("Failed to download audio", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(
			instance.ID,
			data,
			fileName,
			audio.GetMimetype(),
		)

		if err != nil {
			fmt.Println("Failed to save audio", err)
			return mediaType, ""
		}
		fmt.Println("Audio saved")
	}

	image := message.GetImageMessage()
	if image != nil {
		mediaType = "image"
		data, err := instance.Client.Download(image)
		if err != nil {
			fmt.Println("Failed to download image", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(
			instance.ID,
			data,
			fileName,
			image.GetMimetype(),
		)
		if err != nil {
			fmt.Println("Failed to save image", err)
			return mediaType, ""
		}
		fmt.Println("Image saved")
	}

	sticker := message.GetStickerMessage()
	if sticker != nil {
		mediaType = "image"
		data, err := instance.Client.Download(sticker)
		if err != nil {
			fmt.Println("Failed to download sticker", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(
			instance.ID,
			data,
			fileName,
			sticker.GetMimetype(),
		)
		if err != nil {
			fmt.Println("Failed to download sticker", err)
			return mediaType, ""
		}

		fmt.Println("Sticker saved")
	}

	if path != "" && mediaType != "" {
		return mediaType, path
	}

	return "", ""
}

func (m *messageService) getTextMessage(message *waProto.Message) string {
	extendedTextMessage := message.GetExtendedTextMessage()
	if extendedTextMessage != nil {
		return *extendedTextMessage.Text
	}
	return message.GetConversation()
}
