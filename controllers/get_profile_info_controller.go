package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getProfileInfoController struct {
	wppService services.WppService
}

func NewGetProfileInfoController(
	wppService services.WppService,
) *getProfileInfoController {
	return &getProfileInfoController{
		wppService: wppService,
	}
}

// Get Profile Information
// @Summary Get Profile Information
// @Description Retrieves profile information.
// @Tags WhatsApp Profile
// @Param instanceId path string true "Instance ID"
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Profile Information"
// @Router /{instanceId}/profile [get]
func (s *getProfileInfoController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := s.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	jid, ok := utils.MakeJID(instance.Client.Store.ID.User)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}

	info, err := s.wppService.GetContactInfo(instanceID, jid)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Info": info,
	})
}
