package controllers

import (
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
)

type audioMessageBody struct {
	Phone  string
	Base64 string
}

type sendAudioMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
}

type sendAudioMessageResponse struct {
	Message services.Message
}

func NewSendAudioMessageController(
	wppService services.WppService,
	messageService services.MessageService,
) *sendAudioMessageController {
	return &sendAudioMessageController{
		wppService:     wppService,
		messageService: messageService,
	}
}

// Send Audio Message on WhatsApp
// @Summary Send Audio Message on WhatsApp
// @Description Sends an audio message on WhatsApp using the specified instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Param data body audioMessageBody true "Audio message body"
// @Accept json
// @Produce json
// @Success 200 {object} sendAudioMessageResponse "Message Send Response"
// @Router /{instanceId}/chat/send/audio [post]
func (a *sendAudioMessageController) Handler(c *gin.Context) {
	var body audioMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "error trying to validate infos")
		return
	}

	jid, ok := utils.MakeJID(body.Phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}
	instanceID := c.Param("instanceId")

	mimitype, err := utils.GetMimeTypeFromDataURI(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	audioURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	resp, err := a.wppService.SendImageMessage(instanceID, jid, audioURL, mimitype)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	path, err := utils.SaveMedia(
		instanceID,
		resp.ID,
		audioURL.Data,
		mimitype,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		FromMe:     true,
		ChatJID:    jid.User,
		SenderJID:  resp.Sender.User,
		InstanceID: instanceID,
		Timestamp:  resp.Timestamp,
		MessageID:  resp.ID,
		MediaType:  "audio",
		MediaPath:  path,
	}

	err = a.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, sendAudioMessageResponse{
		Message: a.messageService.ToJSON(message),
	})
}
