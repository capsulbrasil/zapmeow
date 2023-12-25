package handler

import (
	"net/http"
	"zapmeow/api/helper"
	"zapmeow/api/response"
	"zapmeow/api/service"
	"zapmeow/pkg/whatsapp"

	"github.com/gin-gonic/gin"
)

type getProfileInfoResponse struct {
	Info whatsapp.ContactInfo `json:"info"`
}

type getProfileInfoHandler struct {
	whatsAppService service.WhatsAppService
}

func NewGetProfileInfoHandler(
	whatsAppService service.WhatsAppService,
) *getProfileInfoHandler {
	return &getProfileInfoHandler{
		whatsAppService: whatsAppService,
	}
}

// Get Profile Information
// @Summary Get Profile Information
// @Description Retrieves profile information.
// @Tags WhatsApp Profile
// @Param instanceId path string true "Instance ID"
// @Accept json
// @Produce json
// @Success 200 {object} getProfileInfoResponse "Profile Information"
// @Router /{instanceId}/profile [get]
func (h *getProfileInfoHandler) Handler(c *gin.Context) {
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

	jid, ok := helper.MakeJID(instance.Client.Store.ID.User)
	if !ok {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")
		return
	}

	info, err := h.whatsAppService.GetContactInfo(instance, jid)
	if err != nil || info == nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, getProfileInfoResponse{
		Info: *info,
	})
}
