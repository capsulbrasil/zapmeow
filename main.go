package main

import (
	"fmt"
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
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("failed load .env")
	}

	config, err := configs.LoadConfigs()
	if err != nil {
		panic("failed get .env")
	}

	// whatsmeow instances
	instances := make(map[string]*configs.Instance)

	// whatsmeow configs
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	whatsmeowContainer, err := sqlstore.New("sqlite3", "file:"+config.DatabaseURL+"?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	databaseClient, err := gorm.Open(sqlite.Open(config.DatabaseURL))
	if err != nil {
		panic("Failed to connect to database")
	}

	db, err := databaseClient.DB()
	if err != nil {
		panic("Failed to get database connection")
	}
	defer db.Close()

	err = databaseClient.AutoMigrate(
		&models.Account{},
		&models.Message{},
	)
	if err != nil {
		panic(err)
	}

	// redis configs
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	stopCh := make(chan struct{})

	// app configs
	app := configs.NewZapMeow(
		whatsmeowContainer,
		databaseClient,
		redisClient,
		instances,
		config,
		&wg,
		stopCh,
	)

	// repositories
	messageRepo := repositories.NewMessageRepository(app.DatabaseClient)
	accountRepo := repositories.NewAccountRepository(app.DatabaseClient)

	// services
	messageService := services.NewMessageService(messageRepo)
	accountService := services.NewAccountService(accountRepo, messageService)
	wppService := services.NewWppService(
		app,
		messageService,
		accountService,
	)

	// workers
	historySyncWorker := workers.NewHistorySyncWorker(
		app,
		messageService,
		accountRepo,
	)

	r := routes.SetupRouter(
		app,
		wppService,
		messageService,
		accountService,
	)

	accounts, err := accountService.GetConnectedAccounts()
	fmt.Println("loading instances...")
	if err != nil {
		fmt.Println("[accounts]: ", err)
	}

	for _, account := range accounts {
		fmt.Println("[instance]: ", account.InstanceID)
		_, err := wppService.GetInstance(account.InstanceID)
		if err != nil {
			fmt.Println("[instance]: ", err)
		}
	}

	go func() {
		if err := r.Run(config.Port); err != nil {
			fmt.Println(err)
		}
	}()

	go historySyncWorker.ProcessQueue()

	<-stopCh

	wg.Wait()
	close(stopCh)
}
