package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getQrCodeController struct {
	wppService     services.WppService
	messageService services.MessageService
	accountService services.AccountService
}

func NewGetQrCodeController(
	wppService services.WppService,
	messageService services.MessageService,
	accountService services.AccountService,
) *getQrCodeController {
	return &getQrCodeController{
		wppService:     wppService,
		messageService: messageService,
		accountService: accountService,
	}
}

func (q *getQrCodeController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	_, err := q.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	account, err := q.accountService.GetAccountByInstanceID(instanceID)
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
