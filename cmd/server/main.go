package main

import (
	"sync"
	"zapmeow/api/model"
	"zapmeow/api/repository"
	"zapmeow/api/route"
	"zapmeow/api/service"
	"zapmeow/config"
	"zapmeow/docs"
	"zapmeow/pkg/database"
	"zapmeow/pkg/logger"
	"zapmeow/pkg/queue"
	"zapmeow/pkg/whatsapp"
	"zapmeow/pkg/zapmeow"
	"zapmeow/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title			ZapMeow API
// @version		1.0
// @description	API to handle multiple WhatsApp instances
// @host			localhost:8900
// @BasePath		/api
func main() {
	docs.SwaggerInfo.BasePath = "/api"

	logger.Init()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading dotfile. ", err)
	}
	cfg := config.Load()

	if cfg.Environment == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	var instances sync.Map // whatsmeow instances
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	stopCh := make(chan struct{})

	whatsApp := whatsapp.NewWhatsApp(cfg.DatabaseURL)
	queue := queue.NewQueue(cfg.RedisAddr, cfg.RedisPassword)

	database := database.NewDatabase(cfg.DatabaseURL)
	err = database.RunMigrate(
		&model.Account{},
		&model.Message{},
	)
	if err != nil {
		logger.Fatal("Error when running gorm automigrate. ", err)
	}

	app := zapmeow.NewZapMeow(
		database,
		queue,
		cfg,
		&instances,
		&wg,
		&mutex,
		&stopCh,
	)

	// repository
	messageRepo := repository.NewMessageRepository(app.Database)
	accountRepo := repository.NewAccountRepository(app.Database)

	// service
	messageService := service.NewMessageService(messageRepo)
	accountService := service.NewAccountService(accountRepo, messageService)
	whatsAppService := service.NewWhatsAppService(
		app,
		messageService,
		accountService,
		whatsApp,
	)

	// workers
	historySyncWorker := worker.NewHistorySyncWorker(
		app,
		messageService,
		accountService,
		whatsAppService,
	)

	r := route.SetupRouter(
		app,
		whatsAppService,
		messageService,
		accountService,
	)

	logger.Info("Loading whatsapp instances")
	accounts, err := accountService.GetConnectedAccounts()
	if err != nil {
		logger.Fatal("Error getting accounts. ", err)
	}

	for _, account := range accounts {
		logger.Info("Loading instance: ", account.InstanceID)
		_, err := whatsAppService.GetInstance(account.InstanceID)
		if err != nil {
			logger.Error("Error getting instance. ", err)
		}
	}

	go func() {
		if err := r.Run(cfg.Port); err != nil {
			logger.Fatal(err)
		}
	}()

	go historySyncWorker.ProcessQueue()

	<-*app.StopCh
	app.Wg.Wait()
	close(*app.StopCh)
}
