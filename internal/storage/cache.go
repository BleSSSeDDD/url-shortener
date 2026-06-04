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
	GetFromCache(ctx context.Context, code string) (string, error)
	AddToCache(ctx context.Context, code string, url string) error
}

type cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) Cache {
	return &cache{rdb: rdb}
}

func (cache *cache) AddToCache(parentCtx context.Context, code string, url string) error {
	ctx, cancel := context.WithTimeout(parentCtx, REDIS_TIMEOUT)
	defer cancel()

	if err := cache.rdb.Set(ctx, code, url, REDIS_TTL).Err(); err != nil {
		return err
	}

	return nil
}

func (cache *cache) GetFromCache(parentCtx context.Context, code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(parentCtx, REDIS_TIMEOUT)
	defer cancel()

	response := cache.rdb.Get(ctx, code)

	if response.Err() != nil {
		return "", response.Err()
	}

	return response.Val(), nil
}
