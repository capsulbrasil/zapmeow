package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getStatusController struct {
	wppService     services.WppService
	accountService services.AccountService
}

func NewGetStatusController(
	wppService services.WppService,
	accountService services.AccountService,
) *getStatusController {
	return &getStatusController{
		wppService:     wppService,
		accountService: accountService,
	}
}

// Get WhatsApp Instance Status
// @Summary Get WhatsApp Instance Status
// @Description Returns the status of the specified WhatsApp instance.
// @Tags WhatsApp Status
// @Param instanceId path string true "Instance ID"
// @Accept json
// @Produce json
// @Success 200 {string} string "Status Response"
// @Router /{instanceId}/status [get]
func (s *getStatusController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := s.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	account, err := s.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	if err != nil {
		utils.RespondNotFound(c, "Account not found")
		return
	}

	var status = account.Status
	if !instance.Client.IsConnected() {
		status = "DISCONNECTED"
	}

	if status == "CONNECTED" && !instance.Client.IsLoggedIn() {
		status = "UNPAIRED"
	}

	utils.RespondWithSuccess(c, gin.H{
		"Status": status,
	})
}
