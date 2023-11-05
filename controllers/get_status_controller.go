package controllers

import (
	"zapmeow/configs"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getStatusController struct {
	app            *configs.ZapMeow
	wppService     services.WppService
	accountService services.AccountService
}

type getStatusResponse struct {
	Status string
}

func NewGetStatusController(
	app *configs.ZapMeow,
	wppService services.WppService,
	accountService services.AccountService,
) *getStatusController {
	return &getStatusController{
		app:            app,
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
// @Success 200 {object} getStatusResponse "Status Response"
// @Router /{instanceId}/status [get]
func (s *getStatusController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := s.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	s.app.Mutex.Lock()
	defer s.app.Mutex.Unlock()
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

	utils.RespondWithSuccess(c, getStatusResponse{
		Status: status,
	})
}
