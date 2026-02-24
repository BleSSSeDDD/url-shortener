package service

import (
	"fmt"
	"math/rand"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type UrlShortener interface {
	Get(shortCode string) (originalUrl string, err error)
	Set(url string) (shortenedUrl string, err error)
}

type urlShortener struct {
	cache   storage.Cache
	storage storage.Postgres
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener(cache storage.Cache, storage storage.Postgres) UrlShortener {
	return &urlShortener{storage: storage, cache: cache}
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
func (u *urlShortener) Set(url string) (shortenedUrl string, err error) {
	// Генерируем новый уникальный код
	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()
		existingCode, seterr := u.storage.SetNewPair(url, code)
		if seterr == nil {
			return existingCode, nil
		} else if pgErr, ok := seterr.(*pq.Error); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return existingCode, nil
			}
			return "", seterr
		}
	}

	return "", fmt.Errorf("failed to generate unique code after 10 attempts")
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *urlShortener) Get(shortCode string) (originalUrl string, err error) {
	originalUrl, err = u.cache.GetFromCache(shortCode)
	if err == nil {
		return originalUrl, nil
	}
	originalUrl, err = u.storage.GetUrlFromCode(shortCode)
	if err != nil {
		return "", err
	}
	u.cache.AddToCache(shortCode, originalUrl)
	return originalUrl, err
}
