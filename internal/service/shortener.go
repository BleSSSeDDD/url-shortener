// Package service пакет для бизнес-логики сокращения ссылок
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/BleSSSeDDD/url-shortener/internal/handlers"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"golang.org/x/sync/singleflight"
)

type CacheGetter interface {
	GetFromCache(ctx context.Context, code string) (string, error)
}

type CacheSetter interface {
	AddToCache(ctx context.Context, code string, url string) error
}

type StorageGetter interface {
	GetURLFromCode(ctx context.Context, code string) (originalURL string, err error)
}

type StorageSetter interface {
	SetNewPair(ctx context.Context, url string, code string) (string, error)
}

// Глобальные переменные для построения уникальных ссылок
const (
	CodeLength  = 6
	MaxAttempts = 10
	URLCharset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type urlShortener struct {
	cacheGetter   CacheGetter
	cacheSetter   CacheSetter
	storageGetter StorageGetter
	storageSetter StorageSetter
	group         singleflight.Group
}

// NewURLShortener cоздает структуру UrlShortener, возвращает на неё указатель
func NewURLShortener(cacheGetter CacheGetter, cacheSetter CacheSetter, storageGetter StorageGetter, storageSetter StorageSetter) (handlers.URLShortenerGetter, handlers.URLShortenerSetter) {
	short := &urlShortener{cacheGetter: cacheGetter, cacheSetter: cacheSetter, storageGetter: storageGetter, storageSetter: storageSetter}
	return short, short
}

// Генерирует случайную строку из 6 символов
func generateShortenedURL() string {
	res := make([]byte, CodeLength)

	for i := 0; i < CodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(URLCharset))))
		if err != nil {
			log.Printf("ошибка при генерации криптостойкой последовательности: %v", err)
			res[i] = URLCharset[0]
			continue
		}
		res[i] = URLCharset[n.Int64()]
	}

	return string(res)
}

// Принимает: оригинальный URL
//
// Возвращает: короткий код и ошибку
//
// Логика: генерирует код до тех пор, пока он не будет уникальным, сохраняет в базу, возвращает код
//
// ЛИБО если такое уже есть, то отдаём чё есть
func (u *urlShortener) Set(ctx context.Context, url string) (shortenedURL string, err error) {
	// Генерируем новый уникальный код
	for i := 0; i < MaxAttempts; i++ {
		code := generateShortenedURL()

		existingCode, seterr := u.storageSetter.SetNewPair(ctx, url, code)

		if seterr == nil {
			return existingCode, nil
		} else if pgErr, ok := seterr.(*pq.Error); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return existingCode, nil
			}
			return "", seterr
		}
	}

	return "", fmt.Errorf("не получилось сгенерировать код за %d попыток", MaxAttempts)
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *urlShortener) Get(ctx context.Context, shortCode string) (originalURL string, err error) {
	originalURL, err = u.cacheGetter.GetFromCache(ctx, shortCode)
	if err == nil {
		return originalURL, nil
	}

	code, singleflightError, _ := u.group.Do(shortCode, func() (any, error) {
		return u.storageGetter.GetURLFromCode(ctx, shortCode)
	})

	if singleflightError != nil {
		return "", singleflightError
	}

	originalURL, ok := code.(string)
	if !ok {
		return "", fmt.Errorf("не удалось скастить ответ singleflight до стринга")
	}

	if addToCacheErr := u.cacheSetter.AddToCache(ctx, shortCode, originalURL); addToCacheErr != nil {
		log.Printf("ошибка сохранения в кеш: %v", addToCacheErr)
	}

	return originalURL, nil
}
