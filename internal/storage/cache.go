package storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	REDIS_TIMEOUT = 1 * time.Second
	REDIS_TTL     = 60 * time.Second
)

type Cache interface {
	GetFromCache(code string) (url string, err error)
	AddToCache(code string, url string)
}

type cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) Cache {
	return &cache{rdb: rdb}
}

func (cache *cache) AddToCache(code string, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), REDIS_TIMEOUT)
	defer cancel()
	cache.rdb.Set(ctx, code, url, REDIS_TTL)
}

func (cache *cache) GetFromCache(code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), REDIS_TIMEOUT)
	defer cancel()
	response := cache.rdb.Get(ctx, code)
	if response.Err() != nil {
		return "", response.Err()
	}
	return response.Val(), nil
}
