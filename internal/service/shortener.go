package service

import (
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type UrlShortener struct {
	rdb *redis.Client
	db  *sql.DB
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener(db *sql.DB, rdb *redis.Client) *UrlShortener {
	return &UrlShortener{db: db, rdb: rdb}
}

// Генерирует случайную строку из 6 символов
func generateShortenedUrl() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	res := make([]byte, 6)

	for i := 0; i < 6; i++ {
		res[i] = charset[rand.Intn(len(charset))]
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
func (u *UrlShortener) Set(url string) (shortenedUrl string, err error) {
	// Проверяем существующий URL, если
	if existingCode, err := storage.GetCodeFromUrl(u.db, url); err == nil {
		return existingCode, nil
	}

	// Генерируем новый уникальный код
	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()
		seterr := storage.SetNewPair(u.db, url, code)
		if seterr == nil {
			return code, nil
		} else if pgErr, ok := seterr.(*pq.Error); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				continue
			}
			return "", seterr
		}
	}

	return "", fmt.Errorf("failed to generate unique code after 10 attempts")
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *UrlShortener) Get(shortCode string) (originalUrl string, err error) {
	originalUrl, err = storage.GetFromCache(u.rdb, shortCode)
	if err == nil {
		return originalUrl, nil
	}
	originalUrl, err = storage.GetUrlFromCode(u.db, shortCode)
	if err != nil {
		return "", err
	}
	storage.AddToCache(u.rdb, shortCode, originalUrl)
	return originalUrl, err
}
