package storage

import (
	"context"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/go-redis/redis/v8"
)

// глобальные переменные для работы с ттл и тайм-айтом в запросах к редису
const (
	RedisTineout = 1 * time.Second
	RedisTTL     = 60 * time.Second
)

type cache struct {
	rdb *redis.Client
}

// NewCache для чтения из кеша
func NewCache(rdb *redis.Client) (service.CacheGetter, service.CacheSetter) {
	newCache := &cache{rdb: rdb}
	return newCache, newCache
}

// AddToCache добавляет новую пару ссылка - код в кэш
func (cache *cache) AddToCache(parentCtx context.Context, code string, url string) error {
	ctx, cancel := context.WithTimeout(parentCtx, RedisTineout)
	defer cancel()

	if err := cache.rdb.Set(ctx, code, url, RedisTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (cache *cache) GetFromCache(parentCtx context.Context, code string) (url string, err error) {
	ctx, cancel := context.WithTimeout(parentCtx, RedisTineout)
	defer cancel()

	response := cache.rdb.Get(ctx, code)

	if response.Err() != nil {
		return "", response.Err()
	}

	return response.Val(), nil
}
