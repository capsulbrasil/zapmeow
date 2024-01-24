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

type sendDocumentMessageBody struct {
	Phone    string `json:"phone"`
	Base64   string `json:"base64"`
	Filename string `json:"filename"`
}

type sendDocumentMessageResponse struct {
	Message response.Message `json:"message"`
}

type sendDocumentMessageHandler struct {
	whatsAppService service.WhatsAppService
	messageService  service.MessageService
}

func NewSendDocumentMessageHandler(
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
) *sendDocumentMessageHandler {
	return &sendDocumentMessageHandler{
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
//	@Param			data		body	sendDocumentMessageBody	true	"Image message body"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sendDocumentMessageResponse	"Message Send Response"
//	@Router			/{instanceId}/chat/send/image [post]
func (h *sendDocumentMessageHandler) Handler(c *gin.Context) {
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

	var body sendDocumentMessageBody
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

	documentURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp, err := h.whatsAppService.SendDocumentMessage(instance, jid, documentURL, mimitype, body.Filename)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	path, err := helper.SaveMedia(
		instanceID,
		resp.ID,
		documentURL.Data,
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
		MediaType:  "document",
		MediaPath:  path,
	}

	err = h.messageService.CreateMessage(&message)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, sendDocumentMessageResponse{
		Message: response.NewMessageResponse(message),
	})
}
