package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	GetFromCache(code string) (url string, err error)
	AddToCache(code string, url string)
}

type cache struct {
	rdb *redis.Client
}

func CacheInit() (Cache, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	var redisConnectErr error

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

	return &cache{rdb: rdb}, redisConnectErr
}

func (cache *cache) AddToCache(code string, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	cache.rdb.Set(ctx, code, url, 60*time.Second)
}

func (cache *cache) GetFromCache(code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	response := cache.rdb.Get(ctx, code)
	if response.Err() != nil {
		return "", response.Err()
	}
	return response.Val(), nil
}
