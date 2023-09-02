package routes

import (
	"zapmeow/configs"
	"zapmeow/controllers"
	"zapmeow/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	app *configs.App,
	wppService services.WppService,
	messageService services.MessageService,
	accountService services.AccountService,
) *gin.Engine {
	router := gin.Default()

	whatsappController := controllers.NewWhatsAppController(
		app,
		wppService,
		messageService,
		accountService,
	)

	group := router.Group("/api")

	group.GET("/:instanceId/qrcode", whatsappController.GetQrcode)
	group.GET("/:instanceId/status", whatsappController.GetStatus)
	group.POST("/:instanceId/check/phones", whatsappController.CheckPhones)
	group.POST("/:instanceId/chat/messages", whatsappController.GetMessages)
	group.POST("/:instanceId/chat/send/text", whatsappController.SendTextMessage)
	group.POST("/:instanceId/chat/send/image", whatsappController.SendImageMessage)
	group.POST("/:instanceId/chat/send/audio", whatsappController.SendAudioMessage)

	return router
}
