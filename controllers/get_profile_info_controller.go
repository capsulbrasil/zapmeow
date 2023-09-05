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

func (s *getProfileInfoController) Handler(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := s.wppService.GetInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	info, err := s.wppService.GetContactInfo(instanceID, *instance.Client.Store.ID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Info": info,
	})
}
