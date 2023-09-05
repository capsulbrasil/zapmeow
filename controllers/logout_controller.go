package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type logoutController struct {
	wppService     services.WppService
	accountService services.AccountService
}

func NewLogoutController(
	wppService services.WppService,
	accountService services.AccountService,
) *logoutController {
	return &logoutController{
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

	utils.RespondWithSuccess(c, gin.H{})
}
