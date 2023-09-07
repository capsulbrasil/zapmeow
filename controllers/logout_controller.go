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

// Logout from WhatsApp
// @Summary Logout from WhatsApp
// @Description Logs out from the specified WhatsApp instance.
// @Tags WhatsApp Logout
// @Param instanceId path string true "Instance ID"
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Router /{instanceId}/logout [post]
func (s *logoutController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	err := s.wppService.Logout(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}
	utils.RespondWithSuccess(c, gin.H{})
}
