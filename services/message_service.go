package services

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"
	"zapmeow/models"
	"zapmeow/utils"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (m *MessageService) downloadMessageMedia(message *waProto.Message, instance *whatsmeow.Client, fileName string) (string, string) {
	path := ""
	mediaType := ""

	dir, err := utils.MakeUserDirectory(instance.Store.ID.User)
	if err != nil {
		return "", ""
	}

	document := message.GetDocumentMessage()
	if document != nil {
		mediaType = "document"

		data, err := instance.Download(document)

		if err != nil {
			fmt.Println("Failed to download document", err)
			return mediaType, ""
		}

		extension := ""
		exts, err := mime.ExtensionsByType(document.GetMimetype())
		if err != nil {
			extension = exts[0]
		} else {
			filename := document.FileName
			extension = filepath.Ext(*filename)
		}

		path, err = utils.SaveMedia(data, dir, fileName, extension)
		if err != nil {
			fmt.Println("Failed to save document", err)
			return mediaType, ""
		}
		fmt.Println("Document saved")
	}

	audio := message.GetAudioMessage()
	if audio != nil {
		mediaType = "audio"

		data, err := instance.Download(audio)
		if err != nil {
			fmt.Println("Failed to download audio", err)
			return mediaType, ""
		}

		exts, err := mime.ExtensionsByType(audio.GetMimetype())
		if err != nil {
			fmt.Println("Failed to get mimetype", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(data, dir, fileName, exts[0])
		if err != nil {
			fmt.Println("Failed to save audio", err)
			return mediaType, ""
		}
		fmt.Println("Audio saved")
	}

	image := message.GetImageMessage()
	if image != nil {
		mediaType = "image"
		data, err := instance.Download(image)
		if err != nil {
			fmt.Println("Failed to download image", err)
			return mediaType, ""
		}

		exts, err := mime.ExtensionsByType(image.GetMimetype())
		if err != nil {
			fmt.Println("Failed to get mimetype", err)
			return mediaType, ""
		}

		path, err = utils.SaveMedia(data, dir, fileName, exts[0])
		if err != nil {
			fmt.Println("Failed to save image", err)
			return mediaType, ""
		}
		fmt.Println("Image saved")
	}

	sticker := message.GetStickerMessage()
	if sticker != nil {
		mediaType = "image"
		data, err := instance.Download(sticker)
		if err != nil {
			fmt.Println("Failed to download sticker", err)
			return mediaType, ""
		}

		exts, err := mime.ExtensionsByType(sticker.GetMimetype())
		if err != nil {
			return mediaType, ""
		}

		path, err = utils.SaveMedia(data, dir, fileName, exts[0])
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

func (m *MessageService) getTextMessage(message *waProto.Message) string {
	extendedTextMessage := message.GetExtendedTextMessage()
	if extendedTextMessage != nil {
		return *extendedTextMessage.Text
	}
	return message.GetConversation()
}

func (m *MessageService) Parse(instance *whatsmeow.Client, msg *events.Message) *models.Message {
	mediaType, path := m.downloadMessageMedia(
		msg.Message,
		instance,
		msg.Info.ID,
	)

	var body = m.getTextMessage(msg.Message)
	if mediaType == "" && body == "" {
		return nil
	}

	if mediaType != "" {
		return &models.Message{
			MeJID:     instance.Store.ID.User,
			MessageID: msg.Info.ID,
			FromMe:    msg.Info.MessageSource.IsFromMe,
			ChatJID:   msg.Info.Chat.User,
			SenderJID: msg.Info.Sender.User,
			Body:      body,
			MediaPath: mediaType,
			MediaType: path,
			Timestamp: msg.Info.Timestamp,
		}
	}

	return &models.Message{
		MeJID:     instance.Store.ID.User,
		MessageID: msg.Info.ID,
		FromMe:    msg.Info.MessageSource.IsFromMe,
		ChatJID:   msg.Info.Chat.User,
		SenderJID: msg.Info.Sender.User,
		Body:      body,
		Timestamp: msg.Info.Timestamp,
	}
}

func (m *MessageService) ToJSON(message models.Message) map[string]interface{} {
	messageJson := map[string]interface{}{
		"ID":        message.ID,
		"Sender":    message.SenderJID,
		"Chat":      message.ChatJID,
		"Me":        message.MeJID,
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
