package handler

import (
	"net/http"
	"zapmeow/api/helper"
	"zapmeow/api/model"
	"zapmeow/api/response"
	"zapmeow/api/service"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
)

type sendImageMessageBody struct {
	Phone  string `json:"phone"`
	Base64 string `json:"base64"`
}

type sendImageMessageResponse struct {
	Message response.Message `json:"message"`
}

type sendImageMessageHandler struct {
	whatsAppService service.WhatsAppService
	messageService  service.MessageService
}

func NewSendImageMessageHandler(
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
) *sendImageMessageHandler {
	return &sendImageMessageHandler{
		whatsAppService: whatsAppService,
		messageService:  messageService,
	}
}

// Send Image Message on WhatsApp
//
//	@Summary		Send Image Message on WhatsApp
//	@Description	Sends an image message on WhatsApp using the specified instance.
//	@Tags			WhatsApp Chat
//	@Param			instanceId	path	string					true	"Instance ID"
//	@Param			data		body	sendImageMessageBody	true	"Image message body"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sendImageMessageResponse	"Message Send Response"
//	@Router			/{instanceId}/chat/send/image [post]
func (h *sendImageMessageHandler) Handler(c *gin.Context) {
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

	var body sendImageMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")
		return
	}

	jid, ok := helper.MakeJID(body.Phone)
	if !ok {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid phone")
		return
	}

	mimitype, err := helper.GetMimeTypeFromDataURI(body.Base64)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	imageURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp, err := h.whatsAppService.SendImageMessage(instance, jid, imageURL, mimitype)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	path, err := helper.SaveMedia(
		instanceID,
		resp.ID,
		imageURL.Data,
		mimitype,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	message := model.Message{
		FromMe:     true,
		ChatJID:    jid.User,
		SenderJID:  resp.Sender.User,
		InstanceID: instanceID,
		Timestamp:  resp.Timestamp,
		MessageID:  resp.ID,
		MediaType:  "image",
		MediaPath:  path,
	}

	err = h.messageService.CreateMessage(&message)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, sendImageMessageResponse{
		Message: response.NewMessageResponse(message),
	})
}
