package handler

import (
	"net/http"
	"zapmeow/api/response"
	"zapmeow/api/service"
	"zapmeow/pkg/whatsapp"

	"github.com/gin-gonic/gin"
)

type getCheckPhonesBody struct {
	Phones []string `json:"phones"`
}

type getCheckPhonesResponse struct {
	Phones []whatsapp.IsOnWhatsAppResponse `json:"phones"`
}

type checkPhonesHandler struct {
	whatsAppService service.WhatsAppService
}

func NewCheckPhonesHandler(
	whatsAppService service.WhatsAppService,
) *checkPhonesHandler {
	return &checkPhonesHandler{
		whatsAppService: whatsAppService,
	}
}

// Check Phones on WhatsApp
//
//	@Summary		Check Phones on WhatsApp
//	@Description	Verifies if the phone numbers in the provided list are registered WhatsApp users.
//	@Tags			WhatsApp Phone Verification
//	@Param			instanceId	path	string				true	"Instance ID"
//	@Param			data		body	getCheckPhonesBody	true	"Phone list"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	getCheckPhonesResponse	"List of verified numbers"
//	@Router			/{instanceId}/check/phones [post]
func (h *checkPhonesHandler) Handler(c *gin.Context) {
	var body getCheckPhonesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Error trying to validate infos. ")
		return
	}

	instanceID := c.Param("instanceId")
	instance, err := h.whatsAppService.GetInstance(instanceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	phones, err := h.whatsAppService.IsOnWhatsApp(instance, body.Phones)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Response(c, http.StatusOK, getCheckPhonesResponse{
		Phones: phones,
	})
}
