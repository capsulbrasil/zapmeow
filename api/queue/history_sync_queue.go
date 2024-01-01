package queue

import (
	"encoding/json"
	"zapmeow/pkg/logger"
	"zapmeow/pkg/zapmeow"
)

type HistorySyncQueueData struct {
	InstanceID string
	History    []byte
}

type historySyncQueue struct {
	app *zapmeow.ZapMeow
}

type HistorySyncQueue interface {
	Enqueue(item HistorySyncQueueData) error
	Dequeue() (*HistorySyncQueueData, error)
}

func NewHistorySyncQueue(app *zapmeow.ZapMeow) *historySyncQueue {
	return &historySyncQueue{
		app: app,
	}
}

func (q *historySyncQueue) Enqueue(item HistorySyncQueueData) error {
	jsonData, err := json.Marshal(item)
	if err != nil {
		logger.Error("Error enqueue history sync.", logger.Fields{
			"error": err,
		})
		return err
	}

	return q.app.Queue.Enqueue(q.app.Config.HistorySyncQueueName, jsonData)
}

func (q *historySyncQueue) Dequeue() (*HistorySyncQueueData, error) {
	result, err := q.app.Queue.Dequeue(q.app.Config.HistorySyncQueueName)
	if err != nil {
		logger.Error("Error dequeuing history sync", logger.Fields{
			"error": err,
		})
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	var data HistorySyncQueueData
	err = json.Unmarshal(result, &data)
	if err != nil {
		logger.Error("Error unmarshal history sync.", logger.Fields{
			"error": err,
		})
		return nil, err
	}

	return &data, nil
}
