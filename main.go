package main

import (
	"sync"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/repositories"
	"zapmeow/routes"
	"zapmeow/services"
	"zapmeow/workers"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// @title ZapMeow API
// @version 1.0
// @description API to handle multiple WhatsApp instances
// @host localhost:8900
// @BasePath /api
func main() {
	log := configs.NewLogger()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading dotfile. ", err)
	}
	config := configs.LoadConfigs()

	// whatsmeow instances
	var instances sync.Map

	// whatsmeow configs
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	whatsmeowContainer, err := sqlstore.New("sqlite3", "file:"+config.DatabaseURL+"?_foreign_keys=on", dbLog)
	if err != nil {
		log.Fatal("Error loading sqlite whatsmeow container. ", err)
	}

	databaseClient, err := gorm.Open(sqlite.Open(config.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Error creating gorm database. ", err)
	}

	db, err := databaseClient.DB()
	if err != nil {
		log.Fatal("Error getting gorm database. ", err)
	}
	defer db.Close()

	err = databaseClient.AutoMigrate(
		&models.Account{},
		&models.Message{},
	)
	if err != nil {
		log.Fatal("Error when running gorm automigrate. ", err)
	}

	// redis configs
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatal("Error when pinging redis. ", err)
	}

	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	stopCh := make(chan struct{})

	// app configs
	app := configs.NewZapMeow(
		whatsmeowContainer,
		databaseClient,
		redisClient,
		&instances,
		config,
		&wg,
		&mutex,
		&stopCh,
	)

	// repositories
	messageRepo := repositories.NewMessageRepository(app.DatabaseClient)
	accountRepo := repositories.NewAccountRepository(app.DatabaseClient)

	// services
	messageService := services.NewMessageService(messageRepo, log)
	accountService := services.NewAccountService(accountRepo, messageService)
	wppService := services.NewWppService(
		app,
		messageService,
		accountService,
		log,
	)

	// workers
	historySyncWorker := workers.NewHistorySyncWorker(
		app,
		messageService,
		accountService,
		wppService,
		log,
	)

	r := routes.SetupRouter(
		app,
		wppService,
		messageService,
		accountService,
	)

	log.Info("Loading whatsapp instances")
	accounts, err := accountService.GetConnectedAccounts()
	if err != nil {
		log.Fatal("Error getting accounts. ", err)
	}

	for _, account := range accounts {
		log.Info("Loading instance: ", account.InstanceID)
		_, err := wppService.GetInstance(account.InstanceID)
		if err != nil {
			log.Error("Error getting instance. ", err)
		}
	}

	go func() {
		if err := r.Run(config.Port); err != nil {
			log.Fatal(err)
		}
	}()

	go historySyncWorker.ProcessQueue()

	<-*app.StopCh

	app.Wg.Wait()
	close(*app.StopCh)
}
