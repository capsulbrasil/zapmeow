package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getContactInfoController struct {
	wppService services.WppService
}

type contactInfoResponse struct {
	Info services.ContactInfo
}

func NewGetContactInfoController(
	wppService services.WppService,
) *getContactInfoController {
	return &getContactInfoController{
		wppService: wppService,
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
func (s *getContactInfoController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")
	phone := c.Query("phone")

	jid, ok := utils.MakeJID(phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}

	info, err := s.wppService.GetContactInfo(instanceID, jid)
	if err != nil || info == nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, contactInfoResponse{
		Info: *info,
	})
}
