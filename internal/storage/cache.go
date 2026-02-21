package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func CacheInit() (rdb *redis.Client, redisConnectErr error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		redisConnectErr = rdb.Ping(ctx).Err()
		cancel()
		if redisConnectErr != nil {
			fmt.Println("Подключение к редису не прошло, пробуем...")
			time.Sleep(time.Second)
		} else {
			fmt.Println("Подключились к редису")
			break
		}
	}

	if redisConnectErr != nil {
		return nil, redisConnectErr
	}

	return rdb, redisConnectErr
}
