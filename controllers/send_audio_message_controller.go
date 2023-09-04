package controllers

import (
	"context"
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type audioMessageBody struct {
	Phone  string
	Base64 string
}
type sendAudioMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
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
// @Success 200 {object} map[string]interface{} "Message Send Response"
// @Router /{instanceId}/chat/send/audio [post]
func (a *sendAudioMessageController) Handler(c *gin.Context) {
	var body audioMessageBody
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

	instance, err := a.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	audioURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	uploaded, err := instance.Client.Upload(
		context.Background(),
		audioURL.Data,
		whatsmeow.MediaAudio,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			Ptt:           proto.Bool(true),
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimitype),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(audioURL.Data))),
		},
	}

	resp, err := instance.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	path, err := utils.SaveMedia(
		instanceID,
		audioURL.Data,
		resp.ID,
		mimitype,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:   jid.User,
		SenderJID: instance.Client.Store.ID.User,
		MediaType: "audio",
		MediaPath: path,
		Timestamp: resp.Timestamp,
		FromMe:    true,
		MessageID: resp.ID,
	}

	err = a.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": a.messageService.ToJSON(message),
	})
}
