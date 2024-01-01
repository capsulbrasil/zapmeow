package route

import (
	"zapmeow/api/handler"
	"zapmeow/api/service"
	"zapmeow/pkg/zapmeow"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	app *zapmeow.ZapMeow,
	whatsAppService service.WhatsAppService,
	messageService service.MessageService,
	accountService service.AccountService,
) *gin.Engine {
	router := gin.Default()

	getQrCodeHandler := handler.NewGetQrCodeHandler(
		app,
		whatsAppService,
		messageService,
		accountService,
	)
	logoutHandler := handler.NewLogoutHandler(
		app,
		whatsAppService,
		accountService,
	)
	getStatusHandler := handler.NewGetStatusHandler(
		app,
		whatsAppService,
		accountService,
	)
	getProfileInfoHandler := handler.NewGetProfileInfoHandler(
		whatsAppService,
	)
	getContactInfoHandler := handler.NewGetContactInfoHandler(
		whatsAppService,
	)
	checkPhonesHandler := handler.NewCheckPhonesHandler(
		whatsAppService,
	)
	getMessagesHandler := handler.NewGetMessagesHandler(
		whatsAppService,
		messageService,
	)
	sendTextMessageHandler := handler.NewSendTextMessageHandler(
		whatsAppService,
		messageService,
	)
	sendImageMessageHandler := handler.NewSendImageMessageHandler(
		whatsAppService,
		messageService,
	)
	sendAudioMessageHandler := handler.NewSendAudioMessageHandler(
		whatsAppService,
		messageService,
	)

	group := router.Group("/api")

	group.GET("/:instanceId/qrcode", getQrCodeHandler.Handler)
	group.GET("/:instanceId/status", getStatusHandler.Handler)
	group.GET("/:instanceId/profile", getProfileInfoHandler.Handler)
	group.GET("/:instanceId/contact/info", getContactInfoHandler.Handler)
	group.POST("/:instanceId/logout", logoutHandler.Handler)
	group.POST("/:instanceId/check/phones", checkPhonesHandler.Handler)
	group.POST("/:instanceId/chat/messages", getMessagesHandler.Handler)
	group.POST("/:instanceId/chat/send/text", sendTextMessageHandler.Handler)
	group.POST("/:instanceId/chat/send/image", sendImageMessageHandler.Handler)
	group.POST("/:instanceId/chat/send/audio", sendAudioMessageHandler.Handler)
	group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}
