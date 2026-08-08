package storage

import (
	"context"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/go-redis/redis/v8"
)

// TTL and timeout used for Redis requests.
const (
	RedisTimeout = 1 * time.Second
	RedisTTL     = 60 * time.Second
)

type cache struct {
	rdb *redis.Client
}

// NewCache builds the cache reader.
func NewCache(rdb *redis.Client) (service.CacheGetter, service.CacheSetter) {
	newCache := &cache{rdb: rdb}
	return newCache, newCache
}

// AddToCache stores a url/code pair in the cache.
func (cache *cache) AddToCache(parentCtx context.Context, code string, url string) error {
	ctx, cancel := context.WithTimeout(parentCtx, RedisTimeout)
	defer cancel()

	if err := cache.rdb.Set(ctx, code, url, RedisTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (cache *cache) GetFromCache(parentCtx context.Context, code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(parentCtx, RedisTimeout)
	defer cancel()

	response := cache.rdb.Get(ctx, code)

	if response.Err() != nil {
		return "", response.Err()
	}

	return response.Val(), nil
}
