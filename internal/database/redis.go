package database

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func CacheInit(addr string) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	var redisConnectErr error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	redisConnectErr = rdb.Ping(ctx).Err()
	cancel()
	if redisConnectErr != nil {
		log.Println("Подключение к редису не прошло")
		return nil, redisConnectErr
	}
	log.Println("Подключились к редису")

	return rdb, redisConnectErr
}
