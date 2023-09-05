package routes

import (
	"zapmeow/configs"
	"zapmeow/controllers"
	"zapmeow/services"

	docs "zapmeow/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	app *configs.ZapMeow,
	wppService services.WppService,
	messageService services.MessageService,
	accountService services.AccountService,
) *gin.Engine {
	docs.SwaggerInfo.BasePath = "/api"

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
	getProfileInfoController := controllers.NewGetProfileInfoController(
		wppService,
	)
	getContactInfoController := controllers.NewGetContactInfoController(
		wppService,
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
	group.GET("/:instanceId/profile", getProfileInfoController.Handler)
	group.POST("/:instanceId/contact/info", getContactInfoController.Handler)
	group.POST("/:instanceId/check/phones", checkPhonesController.Handler)
	group.POST("/:instanceId/chat/messages", getMessagesController.Handler)
	group.POST("/:instanceId/chat/send/text", sendTextMessageController.Handler)
	group.POST("/:instanceId/chat/send/image", sendImageMessageController.Handler)
	group.POST("/:instanceId/chat/send/audio", sendAudioMessageController.Handler)
	group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}
