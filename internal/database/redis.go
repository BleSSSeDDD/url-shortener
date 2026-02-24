package database

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const redisPingTimeout = 1 * time.Second

func CacheInit(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	log.Printf("Attempting to connect to Redis at %s", addr)
	var redisConnectErr error
	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	redisConnectErr = rdb.Ping(ctx).Err()
	cancel()
	if redisConnectErr != nil {
		log.Println("Подключение к редису не прошло")
		return nil, redisConnectErr
	}
	log.Println("Подключились к редису")

	return rdb, redisConnectErr
}
