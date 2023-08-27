package configs

import "os"

type ZapMeowConfig struct {
	StoragePath   string
	WebhookURL    string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	Port          string
	QueueName     string
	MessageLimit  int
}

func LoadConfigs() (ZapMeowConfig, error) {
	storagePath := os.Getenv("STORAGE_PATH")
	webhookURL := os.Getenv("WEBHOOK_URL")
	databaseURL := os.Getenv("DATABASE_PATH")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("PORT")

	return ZapMeowConfig{
		StoragePath:   storagePath,
		WebhookURL:    webhookURL,
		DatabaseURL:   databaseURL,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		Port:          port,
		QueueName:     "HISTORY_SYNC_QUEUE",
		MessageLimit:  10,
	}, nil
}
