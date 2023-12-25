package zapmeow

import (
	"sync"
	"zapmeow/config"
	"zapmeow/pkg/database"
	"zapmeow/pkg/queue"
	"zapmeow/pkg/whatsapp"
)

type ZapMeow struct {
	Database  database.Database
	Queue     queue.Queue
	Config    config.Config
	Instances *sync.Map
	Wg        *sync.WaitGroup
	Mutex     *sync.Mutex
	StopCh    *chan struct{}
}

func NewZapMeow(
	database database.Database,
	queue queue.Queue,
	config config.Config,
	instances *sync.Map,
	wg *sync.WaitGroup,
	mutex *sync.Mutex,
	stopCh *chan struct{},
) *ZapMeow {
	return &ZapMeow{
		Database:  database,
		Queue:     queue,
		Instances: instances,
		Config:    config,
		Wg:        wg,
		Mutex:     mutex,
		StopCh:    stopCh,
	}
}

func (a *ZapMeow) LoadInstance(instanceID string) *whatsapp.Instance {
	value, _ := a.Instances.Load(instanceID)
	if value != nil {
		return value.(*whatsapp.Instance)
	}
	return nil
}

func (a *ZapMeow) StoreInstance(instanceID string, instance *whatsapp.Instance) {
	a.Instances.Store(instanceID, instance)
}

func (a *ZapMeow) DeleteInstance(instanceID string) {
	a.Instances.Delete(instanceID)
}
