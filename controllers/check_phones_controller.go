package controllers

import (
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
)

type phoneCheckBody struct {
	Phones []string
}

type phone struct {
	Query        string
	IsRegistered bool
	JID          struct {
		AD     bool
		User   string
		Agent  uint8
		Device uint8
		Server string
	}
}

type checkPhonesResponse struct {
	Phones []phone
}

type checkPhonesController struct {
	wppService services.WppService
}

func NewCheckPhonesController(
	wppService services.WppService,
) *checkPhonesController {
	return &checkPhonesController{
		wppService: wppService,
	}
}

// Check Phones on WhatsApp
// @Summary Check Phones on WhatsApp
// @Description Verifies if the phone numbers in the provided list are registered WhatsApp users.
// @Tags WhatsApp Phone Verification
// @Param instanceId path string true "Instance ID"
// @Param data body phoneCheckBody true "Phone list"
// @Accept json
// @Produce json
// @Success 200 {object} checkPhonesResponse "List of verified numbers"
// @Router /{instanceId}/check/phones [post]
func (p *checkPhonesController) Handler(c *gin.Context) {
	var body phoneCheckBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Error trying to validate infos. ")
		return
	}
	instanceID := c.Param("instanceId")

	instance, err := p.wppService.GetAuthenticatedInstance(instanceID)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	phones, err := instance.Client.IsOnWhatsApp(body.Phones)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	data := make([]phone, len(phones))
	for i, p := range phones {
		data[i] = phone{
			Query:        p.Query,
			IsRegistered: p.IsIn,
			JID: struct {
				AD     bool
				User   string
				Agent  uint8
				Device uint8
				Server string
			}{
				AD:     p.JID.AD,
				User:   p.JID.User,
				Agent:  p.JID.Agent,
				Device: p.JID.Device,
				Server: p.JID.Server,
			},
		}
	}

	utils.RespondWithSuccess(c, checkPhonesResponse{
		Phones: data,
	})
}
