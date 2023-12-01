package configs

import "os"

type Environment = uint

const (
	Development Environment = iota
	Production
)

type ZapMeowConfig struct {
	Environment            Environment
	StoragePath            string
	WebhookURL             string
	DatabaseURL            string
	RedisAddr              string
	RedisPassword          string
	Port                   string
	QueueName              string
	MaxMessagesPerInstance int
}

func LoadConfigs() ZapMeowConfig {
	storagePath := os.Getenv("STORAGE_PATH")
	webhookURL := os.Getenv("WEBHOOK_URL")
	databaseURL := os.Getenv("DATABASE_PATH")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("PORT")
	env := getEnvironment()

	return ZapMeowConfig{
		Environment:            env,
		StoragePath:            storagePath,
		WebhookURL:             webhookURL,
		DatabaseURL:            databaseURL,
		RedisAddr:              redisAddr,
		RedisPassword:          redisPassword,
		Port:                   port,
		QueueName:              "HISTORY_SYNC_QUEUE",
		MaxMessagesPerInstance: 10,
	}
}

func getEnvironment() Environment {
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		return Production
	}
	return Development
}
