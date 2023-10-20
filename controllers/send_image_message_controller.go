package controllers

import (
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
)

type imageMessageBody struct {
	Phone  string
	Base64 string
}

type sendImageMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
}

type sendImageMessageResponse struct {
	Message services.Message
}

func NewSendImageMessageController(
	wppService services.WppService,
	messageService services.MessageService,
) *sendImageMessageController {
	return &sendImageMessageController{
		wppService:     wppService,
		messageService: messageService,
	}
}

// Send Image Message on WhatsApp
// @Summary Send Image Message on WhatsApp
// @Description Sends an image message on WhatsApp using the specified instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Param data body imageMessageBody true "Image message body"
// @Accept json
// @Produce json
// @Success 200 {object} sendImageMessageResponse "Message Send Response"
// @Router /{instanceId}/chat/send/image [post]
func (i *sendImageMessageController) Handler(c *gin.Context) {
	var body imageMessageBody
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

	mimitype, err := utils.GetMimeTypeFromDataURI(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	imageURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	resp, err := i.wppService.SendImageMessage(instanceID, jid, imageURL, mimitype)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	path, err := utils.SaveMedia(
		instanceID,
		imageURL.Data,
		resp.ID,
		mimitype,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:    jid.User,
		SenderJID:  resp.Sender.User,
		InstanceID: instanceID,
		MediaType:  "image",
		MediaPath:  path,
		Timestamp:  resp.Timestamp,
		FromMe:     true,
		MessageID:  resp.ID,
	}

	err = i.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, sendImageMessageResponse{
		Message: i.messageService.ToJSON(message),
	})
}
