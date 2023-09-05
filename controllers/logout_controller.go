package controllers

import (
	"zapmeow/configs"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type logoutController struct {
	app            *configs.ZapMeow
	wppService     services.WppService
	accountService services.AccountService
}

func NewLogoutController(
	app *configs.ZapMeow,
	wppService services.WppService,
	accountService services.AccountService,
) *logoutController {
	return &logoutController{
		app:            app,
		wppService:     wppService,
		accountService: accountService,
	}
}

func (s *logoutController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := s.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	err = instance.Client.Logout()
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	err = s.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"Status": "UNPAIRED",
	})
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	err = s.accountService.DeleteAccountInfos(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	instance.Client.Disconnect()
	delete(s.app.Instances, instanceID)

	utils.RespondWithSuccess(c, gin.H{})
}
