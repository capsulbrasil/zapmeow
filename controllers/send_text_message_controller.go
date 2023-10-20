package controllers

import (
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type textMessageBody struct {
	Phone string
	Text  string
}

type sendTextMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
}

type sendTextMessageResponse struct {
	Message services.Message
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

// Send Text Message on WhatsApp
// @Summary Send Text Message on WhatsApp
// @Description Sends a text message on WhatsApp using the specified instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Param data body textMessageBody true "Text message body"
// @Accept json
// @Produce json
// @Success 200 {object} sendTextMessageResponse "Message Send Response"
// @Router /{instanceId}/chat/send/text [post]
func (t *sendTextMessageController) Handler(c *gin.Context) {
	var body textMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}

	jid, ok := utils.MakeJID(body.Phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}
	instanceID := c.Param("instanceId")

	resp, err := t.wppService.SendTextMessage(instanceID, jid, body.Text)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:    jid.User,
		SenderJID:  resp.Sender.User,
		InstanceID: instanceID,
		Body:       body.Text,
		Timestamp:  resp.Timestamp,
		FromMe:     true,
		MessageID:  resp.ID,
	}

	err = t.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, sendTextMessageResponse{
		Message: t.messageService.ToJSON(message),
	})
}
