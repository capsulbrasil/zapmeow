package handler

import (
	"net/http"
	"zapmeow/api/helper"
	"zapmeow/api/model"
	"zapmeow/api/response"
	"zapmeow/api/service"

	"github.com/gin-gonic/gin"
)

type sendTextMessageBody struct {
	Phone string `json:"phone"`
	Text  string `json:"text"`
}

type sendTextMessageResponse struct {
	Message response.Message `json:"message"`
}

type sendTextMessageHandler struct {
	whatsAppService service.WhatsAppService
	messageService  service.MessageService
}

func NewSendTextMessageHandler(
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
) *sendTextMessageHandler {
	return &sendTextMessageHandler{
		whatsAppService: whatsAppService,
		messageService:  messageService,
	}
}

// Send Text Message on WhatsApp
// @Summary Send Text Message on WhatsApp
// @Description Sends a text message on WhatsApp using the specified instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Param data body sendTextMessageBody true "Text message body"
// @Accept json
// @Produce json
// @Success 200 {object} sendTextMessageResponse "Message Send Response"
// @Router /{instanceId}/chat/send/text [post]
func (h *sendTextMessageHandler) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")
	instance, err := h.whatsAppService.GetInstance(instanceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var body sendTextMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")

		return
	}

	jid, ok := helper.MakeJID(body.Phone)
	if !ok {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid phone")
		return
	}

	resp, err := h.whatsAppService.SendTextMessage(instance, jid, body.Text)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	message := model.Message{
		MessageID:  resp.ID,
		ChatJID:    jid.User,
		SenderJID:  resp.Sender.User,
		InstanceID: instanceID,
		Body:       body.Text,
		Timestamp:  resp.Timestamp,
		FromMe:     true,
	}

	err = h.messageService.CreateMessage(&message)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, sendTextMessageResponse{
		Message: response.NewMessageResponse(message),
	})
}
