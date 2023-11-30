package queues

import (
	"encoding/json"
	"zapmeow/configs"

	"github.com/go-redis/redis"
)

type HistorySyncQueueData struct {
	InstanceID string
	History    []byte
}

type historySyncQueue struct {
	client *redis.Client
	app    *configs.ZapMeow
	log    configs.Logger
}

type HistorySyncQueue interface {
	Enqueue(item HistorySyncQueueData) error
	Dequeue() (*HistorySyncQueueData, error)
}

func NewHistorySyncQueue(app *configs.ZapMeow, log configs.Logger) *historySyncQueue {
	return &historySyncQueue{
		app: app,
		log: log,
	}
}

func (q *historySyncQueue) Enqueue(item HistorySyncQueueData) error {
	jsonData, err := json.Marshal(item)
	if err != nil {
		q.log.Error("Error marshal history sync", err)
		return err
	}

	return q.app.RedisClient.LPush(q.app.Config.QueueName, jsonData).Err()
}

func (q *historySyncQueue) Dequeue() (*HistorySyncQueueData, error) {
	result, err := q.app.RedisClient.RPop(q.app.Config.QueueName).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			q.log.Error("Error dequeuing history sync", err)
			return nil, err
		}
	}

	var data HistorySyncQueueData
	err = json.Unmarshal(result, &data)

	if err != nil {
		q.log.Error("Error unmarshal history sync.", err)
		return nil, err
	}

	return &data, nil
}
