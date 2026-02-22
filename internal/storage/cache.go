package storage

import (
	"context"
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

func NewCache(rdb *redis.Client) Cache {
	return &cache{rdb: rdb}
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
