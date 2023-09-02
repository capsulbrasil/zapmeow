package queues

import (
	"encoding/json"
	"fmt"
	"zapmeow/configs"

	"github.com/go-redis/redis"
)

type HistorySyncQueueData struct {
	InstanceID string
	History    []byte
}

type historySyncQueue struct {
	client *redis.Client
	app    *configs.App
}

func NewHistorySyncQueue(app *configs.App) *historySyncQueue {
	return &historySyncQueue{
		app: app,
	}
}

func (q *historySyncQueue) Enqueue(item HistorySyncQueueData) error {
	jsonData, err := json.Marshal(item)
	if err != nil {
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
			fmt.Printf("Error dequeuing item: %s", err)
			return nil, err
		}
	}

	var data HistorySyncQueueData
	err = json.Unmarshal(result, &data)

	if err != nil {
		fmt.Printf("Error unmarshal item: %s", err)
		return nil, err
	}

	return &data, nil
}
