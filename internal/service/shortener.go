// Package service пакет для бизнес-логики сокращения ссылок
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"golang.org/x/sync/singleflight"
)

// Глобальные переменные для построения уникальных ссылок
const (
	CodeLength  = 6
	MaxAttempts = 10
	URLCharset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// URLShortener интерфейс для мока работы бизнес-логики получения и записи ссылок
type URLShortener interface {
	Get(ctx context.Context, shortCode string) (originalURL string, err error)
	Set(ctx context.Context, url string) (shortenedURL string, err error)
}

type urlShortener struct {
	cache   storage.Cache
	storage storage.Postgres
	group   singleflight.Group
}

// NewURLShortener cоздает структуру UrlShortener, возвращает на неё указатель
func NewURLShortener(cache storage.Cache, storage storage.Postgres) URLShortener {
	return &urlShortener{storage: storage, cache: cache}
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

		existingCode, seterr := u.storage.SetNewPair(ctx, url, code)

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
	originalURL, err = u.cache.GetFromCache(ctx, shortCode)
	if err == nil {
		return originalURL, nil
	}

	code, singleflightError, _ := u.group.Do(shortCode, func() (any, error) {
		return u.storage.GetURLFromCode(ctx, shortCode)
	})

	if singleflightError != nil {
		return "", err
	}

	originalURL, ok := code.(string)
	if !ok {
		return "", fmt.Errorf("не удалось скастить ответ singleflight до стринга")
	}

	if addToCacheErr := u.cache.AddToCache(ctx, shortCode, originalURL); addToCacheErr != nil {
		log.Printf("ошибка сохранения в кеш: %v", addToCacheErr)
	}

	return originalURL, err
}
