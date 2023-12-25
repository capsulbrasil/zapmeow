package handler

import (
	"net/http"
	"zapmeow/api/helper"
	"zapmeow/api/response"
	"zapmeow/api/service"
	"zapmeow/pkg/whatsapp"

	"github.com/gin-gonic/gin"
)

type contactInfoResponse struct {
	Info whatsapp.ContactInfo `json:"info"`
}

type getContactInfoHandler struct {
	whatsAppService service.WhatsAppService
}

func NewGetContactInfoHandler(
	whatsAppService service.WhatsAppService,
) *getContactInfoHandler {
	return &getContactInfoHandler{
		whatsAppService: whatsAppService,
	}
}

// Get Contact Information
// @Summary Get Contact Information
// @Description Retrieves contact information.
// @Tags WhatsApp Contact
// @Param instanceId path string true "Instance ID"
// @Param phone query string true "Phone"
// @Accept json
// @Produce json
// @Success 200 {object} contactInfoResponse "Contact Information"
// @Router /{instanceId}/contact/info [get]
func (h *getContactInfoHandler) Handler(c *gin.Context) {
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

	phone := c.Query("phone")
	jid, ok := helper.MakeJID(phone)
	if !ok {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")
		return
	}

	info, err := h.whatsAppService.GetContactInfo(instance, jid)
	if err != nil || info == nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, contactInfoResponse{
		Info: *info,
	})
}
