package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type getContactInfoController struct {
	wppService services.WppService
}

func NewGetContactInfoController(
	wppService services.WppService,
) *getContactInfoController {
	return &getContactInfoController{
		wppService: wppService,
	}
}

func (s *getContactInfoController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")
	phone := c.Query("phone")

	jid, ok := utils.MakeJID(phone)
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
