package controllers

import (
	"context"
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type sendTextMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
}

func NewSendTextMessageController(
	wppService services.WppService,
	messageService services.MessageService,
) *sendTextMessageController {
	return &sendTextMessageController{
		wppService:     wppService,
		messageService: messageService,
	}
}

func (t *sendTextMessageController) Handler(c *gin.Context) {
	type Body struct {
		Phone string
		Text  string
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}

	jid, ok := utils.MakeJID(body.Phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}
	instanceId := c.Param("instanceId")

	instance, err := t.wppService.GetAuthenticatedInstance(instanceId)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &body.Text,
		},
	}

	resp, err := instance.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:   jid.User,
		SenderJID: instance.Store.ID.User,
		MeJID:     instance.Store.ID.User,
		Body:      body.Text,
		Timestamp: resp.Timestamp,
		FromMe:    true,
		MessageID: resp.ID,
	}

	err = t.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": t.messageService.ToJSON(message),
	})
}
