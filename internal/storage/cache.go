package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func CacheInit() (rdb *redis.Client, redisConnectErr error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		redisConnectErr = rdb.Ping(ctx).Err()
		cancel()
		if redisConnectErr != nil {
			log.Println("Подключение к редису не прошло, пробуем...")
			time.Sleep(time.Second)
		} else {
			log.Println("Подключились к редису")
			break
		}
	}

	if redisConnectErr != nil {
		return nil, redisConnectErr
	}

	return rdb, redisConnectErr
}

func AddToCache(rdb *redis.Client, code string, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	rdb.Set(ctx, code, url, 60*time.Second)
}

func GetFromCache(rdb *redis.Client, code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	response := rdb.Get(ctx, code)
	if response.Err() != nil {
		return "", response.Err()
	}
	return response.Val(), nil
}
