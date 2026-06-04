package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"golang.org/x/sync/singleflight"
)

const (
	CODE_LENGTH  = 6
	MAX_ATTEMPTS = 10
	URL_CHARSET  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type URLShortener interface {
	Get(ctx context.Context, shortCode string) (originalUrl string, err error)
	Set(ctx context.Context, url string) (shortenedUrl string, err error)
}

type urlShortener struct {
	cache   storage.Cache
	storage storage.Postgres
	group   singleflight.Group
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewURLShortener(cache storage.Cache, storage storage.Postgres) URLShortener {
	return &urlShortener{storage: storage, cache: cache}
}

// Генерирует случайную строку из 6 символов
func generateShortenedURL() string {
	res := make([]byte, CODE_LENGTH)

	for i := 0; i < CODE_LENGTH; i++ {
		res[i] = URL_CHARSET[rand.Intn(len(URL_CHARSET))]
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
func (u *urlShortener) Set(ctx context.Context, url string) (shortenedUrl string, err error) {
	// Генерируем новый уникальный код
	for i := 0; i < MAX_ATTEMPTS; i++ {
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

	return "", fmt.Errorf("не получилось сгенерировать код за %d попыток", MAX_ATTEMPTS)
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *urlShortener) Get(ctx context.Context, shortCode string) (originalUrl string, err error) {
	originalUrl, err = u.cache.GetFromCache(ctx, shortCode)
	if err == nil {
		return originalUrl, nil
	}

	code, singleflightError, _ := u.group.Do(shortCode, func() (any, error) {
		return u.storage.GetUrlFromCode(ctx, shortCode)
	})

	if singleflightError != nil {
		return "", err
	}

	originalUrl, ok := code.(string)
	if !ok {
		return "", fmt.Errorf("не удалось скастить ответ singleflight до стринга")
	}

	if err := u.cache.AddToCache(ctx, shortCode, originalUrl); err != nil {
		log.Printf("ошибка сохранения в кеш: %v", err)
	}

	return originalUrl, err
}
