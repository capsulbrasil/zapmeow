package handler

import (
	"net/http"
	"zapmeow/api/response"
	"zapmeow/api/service"
	"zapmeow/pkg/zapmeow"

	"github.com/gin-gonic/gin"
)

type logoutHandler struct {
	app             *zapmeow.ZapMeow
	whatsAppService service.WhatsAppService
	accountService  service.AccountService
}

func NewLogoutHandler(
	app *zapmeow.ZapMeow,
	whatsAppService service.WhatsAppService,
	accountService service.AccountService,
) *logoutHandler {
	return &logoutHandler{
		app:             app,
		whatsAppService: whatsAppService,
		accountService:  accountService,
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
func (h *logoutHandler) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")
	instance, err := h.whatsAppService.GetInstance(instanceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if !h.whatsAppService.IsAuthenticated(instance) {
		response.ErrorResponse(c, http.StatusUnauthorized, "unautenticated")
		return
	}

	err = h.whatsAppService.Logout(instance)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, gin.H{})
}
