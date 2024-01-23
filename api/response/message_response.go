package response

import (
	"encoding/base64"
	"mime"
	"os"
	"path/filepath"
	"time"
	"zapmeow/api/model"
)

type Message struct {
	ID            uint      `json:"id"`
	Sender        string    `json:"sender"`
	Chat          string    `json:"chat"`
	MessageID     string    `json:"message_id"`
	FromMe        bool      `json:"from_me"`
	Timestamp     time.Time `json:"timestamp"`
	Body          string    `json:"body"`
	MediaType     string    `json:"media_type"`
	MediaMimeType string    `json:"media_mimetype"`
	MediaBase64   string    `json:"media_base64"`
}

func NewMessageResponse(msg model.Message) Message {
	data := Message{
		ID:        msg.ID,
		Sender:    msg.SenderJID,
		Chat:      msg.ChatJID,
		MessageID: msg.MessageID,
		FromMe:    msg.FromMe,
		Timestamp: msg.Timestamp,
		Body:      msg.Body,
		MediaType: msg.MediaType,
	}

	if msg.MediaType != "" {
		media, err := os.ReadFile(msg.MediaPath)
		if err != nil {
			// logger.Error("Error reading the file. ", err)
		} else {
			mimetype := mime.TypeByExtension(filepath.Ext(msg.MediaPath))
			base64 := base64.StdEncoding.EncodeToString(media)
			data.MediaMimeType = mimetype
			data.MediaBase64 = base64
		}
	}

	return data
}

func NewMessagesResponse(msgs *[]model.Message) []Message {
	var data []Message
	for _, message := range *msgs {
		data = append(data, NewMessageResponse(message))
	}

	return data
}
