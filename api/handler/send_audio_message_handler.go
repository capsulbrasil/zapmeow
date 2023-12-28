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

type sendAudioMessageBody struct {
	Phone  string `json:"phone"`
	Base64 string `json:"base64"`
}

type sendAudioMessageResponse struct {
	Message response.Message `json:"message"`
}

type sendAudioMessageHandler struct {
	whatsAppService service.WhatsAppService
	messageService  service.MessageService
}

func NewSendAudioMessageHandler(
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
) *sendAudioMessageHandler {
	return &sendAudioMessageHandler{
		whatsAppService: whatsAppService,
		messageService:  messageService,
	}
}

// Send Audio Message on WhatsApp
//
//	@Summary		Send Audio Message on WhatsApp
//	@Description	Sends an audio message on WhatsApp using the specified instance.
//	@Tags			WhatsApp Chat
//	@Param			instanceId	path	string					true	"Instance ID"
//	@Param			data		body	sendAudioMessageBody	true	"Audio message body"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sendAudioMessageResponse	"Message Send Response"
//	@Router			/{instanceId}/chat/send/audio [post]
func (h *sendAudioMessageHandler) Handler(c *gin.Context) {
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

	var body sendAudioMessageBody
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

	audioURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp, err := h.whatsAppService.SendImageMessage(instance, jid, audioURL, mimitype)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	path, err := helper.SaveMedia(
		instanceID,
		resp.ID,
		audioURL.Data,
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
		MediaType:  "audio",
		MediaPath:  path,
	}

	err = h.messageService.CreateMessage(&message)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, sendAudioMessageResponse{
		Message: response.NewMessageResponse(message),
	})
}
