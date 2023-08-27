package configs

import (
	"sync"

	"github.com/go-redis/redis"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"gorm.io/gorm"
)

type App struct {
	WhatsmeowContainer *sqlstore.Container
	DatabaseClient     *gorm.DB
	RedisClient        *redis.Client
	Instances          map[string]*whatsmeow.Client
	Config             ZapMeowConfig
	Wg                 *sync.WaitGroup
	StopCh             <-chan struct{}
}

func NewApp(
	whatsmeowContainer *sqlstore.Container,
	databaseClient *gorm.DB,
	redisClient *redis.Client,
	instances map[string]*whatsmeow.Client,
	config ZapMeowConfig,
	wg *sync.WaitGroup,
	stopCh <-chan struct{},
) *App {
	return &App{
		WhatsmeowContainer: whatsmeowContainer,
		DatabaseClient:     databaseClient,
		RedisClient:        redisClient,
		Instances:          instances,
		Config:             config,
		Wg:                 wg,
		StopCh:             stopCh,
	}
}
