package config

import (
	"log"
	"os"
	"strconv"
)

type Environment = uint

const (
	Development Environment = iota
	Production
)

type Config struct {
	Environment          Environment
	StoragePath          string
	WebhookURL           string
	DatabaseURL          string
	RedisAddr            string
	RedisPassword        string
	Port                 string
	HistorySyncQueueName string
	HistorySync          bool
	MaxMessageSync       int
}

func Load() Config {
	storagePathEnv := os.Getenv("STORAGE_PATH")
	webhookURLEnv := os.Getenv("WEBHOOK_URL")
	databaseURLEnv := os.Getenv("DATABASE_PATH")
	redisAddrEnv := os.Getenv("REDIS_ADDR")
	redisPasswordEnv := os.Getenv("REDIS_PASSWORD")
	portEnv := os.Getenv("PORT")
	historySyncEnv := os.Getenv("HISTORY_SYNC")
	maxMessageSyncEnv := os.Getenv("MAX_MESSAGE_SYNC")
	environment := getEnvironment()

	maxMessageSync, err := strconv.Atoi(maxMessageSyncEnv)
	if err != nil {
		maxMessageSync = 10
	}

	historySync, err := strconv.ParseBool(historySyncEnv)
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Environment:          environment,
		StoragePath:          storagePathEnv,
		WebhookURL:           webhookURLEnv,
		DatabaseURL:          databaseURLEnv,
		RedisAddr:            redisAddrEnv,
		RedisPassword:        redisPasswordEnv,
		Port:                 portEnv,
		HistorySyncQueueName: "queue:history-sync",
		HistorySync:          historySync,
		MaxMessageSync:       maxMessageSync,
	}
}

func getEnvironment() Environment {
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		return Production
	}
	return Development
}
