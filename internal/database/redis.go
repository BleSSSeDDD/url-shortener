package database

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisPingTimeout bounds the startup connectivity check against Redis.
const RedisPingTimeout = 1 * time.Second

// CacheInit opens a new Redis connection.
func CacheInit(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	log.Printf("connecting to redis at %s", addr)

	var redisConnectErr error

	ctx, cancel := context.WithTimeout(context.Background(), RedisPingTimeout)

	redisConnectErr = rdb.Ping(ctx).Err()

	cancel()

	if redisConnectErr != nil {
		log.Println("redis connection failed")
		return nil, redisConnectErr
	}
	log.Println("connected to redis")

	return rdb, redisConnectErr
}
