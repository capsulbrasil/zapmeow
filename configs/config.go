package configs

import "os"

type ZapMeowConfig struct {
	Env                    string
	StoragePath            string
	WebhookURL             string
	DatabaseURL            string
	RedisAddr              string
	RedisPassword          string
	Port                   string
	QueueName              string
	MaxMessagesPerInstance int
}

func LoadConfigs() (ZapMeowConfig, error) {
	env := os.Getenv("ENV")
	storagePath := os.Getenv("STORAGE_PATH")
	webhookURL := os.Getenv("WEBHOOK_URL")
	databaseURL := os.Getenv("DATABASE_PATH")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("PORT")

	return ZapMeowConfig{
		Env:                    env,
		StoragePath:            storagePath,
		WebhookURL:             webhookURL,
		DatabaseURL:            databaseURL,
		RedisAddr:              redisAddr,
		RedisPassword:          redisPassword,
		Port:                   port,
		QueueName:              "HISTORY_SYNC_QUEUE",
		MaxMessagesPerInstance: 10,
	}, nil
}
