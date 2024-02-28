package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderJID  string `gorm:"column:sender_jid"`
	ChatJID    string `gorm:"column:chat_jid"`
	InstanceID string
	MessageID  string
	Timestamp  time.Time
	Body       string
	MediaType  string // text, image, ptt, audio, document
	MediaPath  string
	FromMe     bool
}
