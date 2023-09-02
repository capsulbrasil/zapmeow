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

	getQrCodeController := controllers.NewGetQrCodeController(
		wppService,
		messageService,
		accountService,
	)
	getStatusController := controllers.NewGetStatusController(
		wppService,
		accountService,
	)
	checkPhonesController := controllers.NewCheckPhonesController(
		wppService,
	)
	getMessagesController := controllers.NewGetMessagesController(
		wppService,
		messageService,
	)
	sendTextMessageController := controllers.NewSendTextMessageController(
		wppService,
		messageService,
	)
	sendImageMessageController := controllers.NewSendImageMessageController(
		wppService,
		messageService,
	)
	sendAudioMessageController := controllers.NewSendAudioMessageController(
		wppService,
		messageService,
	)

	group := router.Group("/api")

	group.GET("/:instanceId/qrcode", getQrCodeController.Handler)
	group.GET("/:instanceId/status", getStatusController.Handler)
	group.POST("/:instanceId/check/phones", checkPhonesController.Handler)
	group.POST("/:instanceId/chat/messages", getMessagesController.Handler)
	group.POST("/:instanceId/chat/send/text", sendTextMessageController.Handler)
	group.POST("/:instanceId/chat/send/image", sendImageMessageController.Handler)
	group.POST("/:instanceId/chat/send/audio", sendAudioMessageController.Handler)

	return router
}
