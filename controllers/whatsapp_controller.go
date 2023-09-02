package controllers

import (
	"context"
	"mime"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type whatsAppController struct {
	app            *configs.App
	wppService     services.WppService
	messageService services.MessageService
	accountService services.AccountService
}

func NewWhatsAppController(
	app *configs.App,
	wppService services.WppService,
	messageService services.MessageService,
	accountService services.AccountService,
) *whatsAppController {
	return &whatsAppController{
		wppService:     wppService,
		messageService: messageService,
		accountService: accountService,
		app:            app,
	}
}

func (wc *whatsAppController) GetQrcode(c *gin.Context) {
	instanceID := c.Param("instanceId")

	_, err := wc.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	account, err := wc.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	if account == nil {
		utils.RespondNotFound(c, "Account not found")
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"QrCode": account.QrCode,
	})
}

func (wc *whatsAppController) GetStatus(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := wc.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	account, err := wc.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	if err != nil {
		utils.RespondNotFound(c, "Account not found")
		return
	}

	var status = account.Status
	if !instance.IsConnected() {
		status = "DISCONNECTED"
	}

	if status == "CONNECTED" && !instance.IsLoggedIn() {
		status = "UNPAIRED"
	}

	utils.RespondWithSuccess(c, gin.H{
		"Status": status,
	})
}

func (wc *whatsAppController) CheckPhones(c *gin.Context) {
	type Body struct {
		Phones []string
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}
	instanceID := c.Param("instanceId")

	instance, err := wc.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	phones, err := instance.IsOnWhatsApp(body.Phones)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	data := make([]gin.H, len(phones))
	for i, p := range phones {
		data[i] = gin.H{
			"Query":        p.Query,
			"IsRegistered": p.IsIn,
			"JID": gin.H{
				"AD":     p.JID.AD,
				"User":   p.JID.User,
				"Agent":  p.JID.Agent,
				"Device": p.JID.Device,
				"Server": p.JID.Server,
			},
		}
	}

	utils.RespondWithSuccess(c, gin.H{
		"Phones": data,
	})
}

func (wc *whatsAppController) GetMessages(c *gin.Context) {
	type Body struct {
		Phone string
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}
	instanceID := c.Param("instanceId")

	instance, err := wc.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	messages, err := wc.messageService.GetChatMessages(
		instance.Store.ID.User,
		body.Phone,
	)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	var data = []gin.H{}
	for _, message := range *messages {
		data = append(data, wc.messageService.ToJSON(message))
	}

	utils.RespondWithSuccess(c, gin.H{
		"Messages": data,
	})
}

func (wc *whatsAppController) SendTextMessage(c *gin.Context) {
	type Body struct {
		Phone string
		Text  string
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

	instance, err := wc.wppService.GetAuthenticatedInstance(instanceId)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &body.Text,
		},
	}

	resp, err := instance.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:   jid.User,
		SenderJID: instance.Store.ID.User,
		MeJID:     instance.Store.ID.User,
		Body:      body.Text,
		Timestamp: resp.Timestamp,
		FromMe:    true,
		MessageID: resp.ID,
	}

	err = wc.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": wc.messageService.ToJSON(message),
	})
}

func (wc *whatsAppController) SendImageMessage(c *gin.Context) {
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

	instance, err := wc.wppService.GetAuthenticatedInstance(instanceId)
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

	err = wc.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": wc.messageService.ToJSON(message),
	})
}

func (wc *whatsAppController) SendAudioMessage(c *gin.Context) {
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

	instance, err := wc.wppService.GetAuthenticatedInstance(instanceId)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	audioURL, err := dataurl.DecodeString(body.Base64)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	uploaded, err := instance.Upload(
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

	resp, err := instance.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	dir, err := utils.MakeUserDirectory(instance.Store.ID.User)
	exts, _ := mime.ExtensionsByType(mimitype)
	path, err := utils.SaveMedia(
		audioURL.Data,
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
		MediaType: "audio",
		MediaPath: path,
		Timestamp: resp.Timestamp,
		FromMe:    true,
		MessageID: resp.ID,
	}

	err = wc.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": wc.messageService.ToJSON(message),
	})
}
