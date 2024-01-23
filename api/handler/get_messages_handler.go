package handler

import (
	"net/http"
	"zapmeow/api/response"
	"zapmeow/api/service"

	"github.com/gin-gonic/gin"
)

type getMessagesBody struct {
	Phone string `json:"phone"`
}

type getMessagesResponse struct {
	Messages []response.Message `json:"messages"`
}

type getMessagesHandler struct {
	whatsAppService service.WhatsAppService
	messageService  service.MessageService
}

func NewGetMessagesHandler(
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
) *getMessagesHandler {
	return &getMessagesHandler{
		whatsAppService: whatsAppService,
		messageService:  messageService,
	}
}

// Get WhatsApp Chat Messages
//
//	@Summary		Get WhatsApp Chat Messages
//	@Description	Returns chat messages from the specified WhatsApp instance.
//	@Tags			WhatsApp Chat
//	@Param			instanceId	path	string			true	"Instance ID"
//	@Param			data		body	getMessagesBody	true	"Phone"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	getMessagesResponse	"List of chat messages"
//	@Router			/{instanceId}/chat/messages [post]
func (h *getMessagesHandler) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")
	instance, err := h.whatsAppService.GetInstance(instanceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if !h.whatsAppService.IsAuthenticated(instance) {
		response.ErrorResponse(c, http.StatusUnauthorized, "unautenticated")
		return
	}

	var body getMessagesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")
		return
	}

	messages, err := h.messageService.GetChatMessages(
		instanceID,
		body.Phone,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, getMessagesResponse{
		Messages: response.NewMessagesResponse(messages),
	})
}
