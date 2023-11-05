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
	Instances          *sync.Map
	Config             ZapMeowConfig
	Wg                 *sync.WaitGroup
	Mutex              *sync.Mutex
	StopCh             *chan struct{}
}

func NewZapMeow(
	whatsmeowContainer *sqlstore.Container,
	databaseClient *gorm.DB,
	redisClient *redis.Client,
	instances *sync.Map,
	config ZapMeowConfig,
	wg *sync.WaitGroup,
	mutex *sync.Mutex,
	stopCh *chan struct{},
) *ZapMeow {
	return &ZapMeow{
		WhatsmeowContainer: whatsmeowContainer,
		DatabaseClient:     databaseClient,
		RedisClient:        redisClient,
		Instances:          instances,
		Config:             config,
		Wg:                 wg,
		Mutex:              mutex,
		StopCh:             stopCh,
	}
}

func (a *ZapMeow) LoadInstance(instanceID string) *Instance {
	value, _ := a.Instances.Load(instanceID)
	if value != nil {
		return value.(*Instance)
	}
	return nil
}

func (a *ZapMeow) StoreInstance(instanceID string, instance *Instance) {
	a.Instances.Store(instanceID, instance)
}

func (a *ZapMeow) DeleteInstance(instanceID string) {
	a.Instances.Delete(instanceID)
}
