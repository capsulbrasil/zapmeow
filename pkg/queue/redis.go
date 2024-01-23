package queue

import (
	"zapmeow/pkg/logger"

	"github.com/go-redis/redis"
)

type Queue interface {
	Enqueue(queueName string, data []byte) error
	Dequeue(queueName string) ([]byte, error)
}

type queue struct {
	client *redis.Client
}

func NewQueue(addr string, password string) *queue {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	if _, err := client.Ping().Result(); err != nil {
		logger.Fatal(err)
	}
	return &queue{
		client: client,
	}
}

func (q *queue) Enqueue(queueName string, data []byte) error {
	return q.client.LPush(queueName, data).Err()
}

func (q *queue) Dequeue(queueName string) ([]byte, error) {
	result, err := q.client.LPop(queueName).Bytes()
	if err != nil && err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return result, nil
}
