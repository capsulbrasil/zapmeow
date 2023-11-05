package controllers

import (
	"zapmeow/configs"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getQrCodeController struct {
	app            *configs.ZapMeow
	wppService     services.WppService
	messageService services.MessageService
	accountService services.AccountService
}

type getQrCodeResponse struct {
	QrCode string
}

func NewGetQrCodeController(
	app *configs.ZapMeow,
	wppService services.WppService,
	messageService services.MessageService,
	accountService services.AccountService,
) *getQrCodeController {
	return &getQrCodeController{
		app:            app,
		wppService:     wppService,
		messageService: messageService,
		accountService: accountService,
	}
}

// Get QR Code for WhatsApp Login
// @Summary Get WhatsApp QR Code
// @Description Returns a QR code to initiate WhatsApp login.
// @Tags WhatsApp Login
// @Param instanceId path string true "Instance ID"
// @Produce json
// @Success 200 {object} getQrCodeResponse "QR Code"
// @Router /{instanceId}/qrcode [get]
func (q *getQrCodeController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	_, err := q.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	q.app.Mutex.Lock()
	defer q.app.Mutex.Unlock()
	account, err := q.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	if account == nil {
		utils.RespondNotFound(c, "Account not found")
		return
	}

	utils.RespondWithSuccess(c, getQrCodeResponse{
		QrCode: account.QrCode,
	})
}
