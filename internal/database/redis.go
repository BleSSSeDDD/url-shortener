package database

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const REDIS_PING_TIMEOUT = 1 * time.Second

func CacheInit(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	log.Printf("Пытаемся подключиться к редису по %s", addr)

	var redisConnectErr error

	ctx, cancel := context.WithTimeout(context.Background(), REDIS_PING_TIMEOUT)

	redisConnectErr = rdb.Ping(ctx).Err()

	cancel()

	if redisConnectErr != nil {
		log.Println("Подключение к редису не прошло")
		return nil, redisConnectErr
	}
	log.Println("Подключились к редису")

	return rdb, redisConnectErr
}
