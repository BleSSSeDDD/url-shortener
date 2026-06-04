package storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// глобальные переменные для работы с ттл и тайм-айтом в запросах к редису
const (
	RedisTineout = 1 * time.Second
	RedisTTL     = 60 * time.Second
)

// Cache интерфейс для мока кэша
type Cache interface {
	GetFromCache(ctx context.Context, code string) (string, error)
	AddToCache(ctx context.Context, code string, url string) error
}

type cache struct {
	rdb *redis.Client
}

// NewCache создает стурктуру для работы с редисом
func NewCache(rdb *redis.Client) Cache {
	return &cache{rdb: rdb}
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
