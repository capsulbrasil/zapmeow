package configs

import (
	"sync"

	"github.com/go-redis/redis"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"gorm.io/gorm"
)

type Instance struct {
	ID     string
	Client *whatsmeow.Client
}

type ZapMeow struct {
	WhatsmeowContainer *sqlstore.Container
	DatabaseClient     *gorm.DB
	RedisClient        *redis.Client
	Instances          map[string]*Instance
	Config             ZapMeowConfig
	Wg                 *sync.WaitGroup
	StopCh             *chan struct{}
}

func NewZapMeow(
	whatsmeowContainer *sqlstore.Container,
	databaseClient *gorm.DB,
	redisClient *redis.Client,
	instances map[string]*Instance,
	config ZapMeowConfig,
	wg *sync.WaitGroup,
	stopCh *chan struct{},
) *ZapMeow {
	return &ZapMeow{
		WhatsmeowContainer: whatsmeowContainer,
		DatabaseClient:     databaseClient,
		RedisClient:        redisClient,
		Instances:          instances,
		Config:             config,
		Wg:                 wg,
		StopCh:             stopCh,
	}
}
