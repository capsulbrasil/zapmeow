package controllers

import (
	"context"
	"mime"
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type sendImageMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
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

func (i *sendImageMessageController) Handler(c *gin.Context) {
	type Body struct {
		Phone  string
		Base64 string
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}

	jid, ok := utils.MakeJID(body.Phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}
	instanceId := c.Param("instanceId")

	mimitype, err := utils.GetMimeTypeFromDataURI(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	instance, err := i.wppService.GetAuthenticatedInstance(instanceId)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	imageURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	uploaded, err := instance.Upload(
		context.Background(),
		imageURL.Data,
		whatsmeow.MediaImage,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimitype),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageURL.Data))),
		},
	}

	resp, err := instance.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	dir, err := utils.MakeUserDirectory(instance.Store.ID.User)
	exts, _ := mime.ExtensionsByType(mimitype)

	path, err := utils.SaveMedia(
		imageURL.Data,
		dir,
		resp.ID,
		exts[0],
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:   jid.User,
		SenderJID: instance.Store.ID.User,
		MeJID:     instance.Store.ID.User,
		MediaType: "image",
		MediaPath: path,
		Timestamp: resp.Timestamp,
		FromMe:    true,
		MessageID: resp.ID,
	}

	err = i.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": i.messageService.ToJSON(message),
	})
}
