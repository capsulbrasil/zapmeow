package utils

import (
	"github.com/go-redis/redis"
)

func Ping(client *redis.Client) error {
	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	return nil
}
