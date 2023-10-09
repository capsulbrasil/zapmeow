package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getMessagesController struct {
	wppService     services.WppService
	messageService services.MessageService
}

type getMessagesResponse struct {
	Messages []services.Message
}

func NewGetMessagesController(
	wppService services.WppService,
	messageService services.MessageService,
) *getMessagesController {
	return &getMessagesController{
		wppService:     wppService,
		messageService: messageService,
	}
}

// Get WhatsApp Chat Messages
// @Summary Get WhatsApp Chat Messages
// @Description Returns chat messages from the specified WhatsApp instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Accept json
// @Produce json
// @Success 200 {object} getMessagesResponse "List of chat messages"
// @Router /{instanceId}/chat/messages [post]
func (m *getMessagesController) Handler(c *gin.Context) {
	type Body struct {
		Phone string
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}
	instanceID := c.Param("instanceId")

	_, err := m.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	messages, err := m.messageService.GetChatMessages(
		instanceID,
		body.Phone,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	var data []services.Message
	for _, message := range *messages {
		data = append(data, m.messageService.ToJSON(message))
	}

	utils.RespondWithSuccess(c, getMessagesResponse{
		Messages: data,
	})
}
