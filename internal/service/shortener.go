package service

import (
	"fmt"
	"math/rand"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

const (
	CODE_LENGTH  = 6
	MAX_ATTEMPTS = 10
	URL_CHARSET  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
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
func (u *urlShortener) Set(url string) (shortenedUrl string, err error) {
	// Генерируем новый уникальный код
	for i := 0; i < MAX_ATTEMPTS; i++ {
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

	return "", fmt.Errorf("failed to generate unique code after %d attempts", MAX_ATTEMPTS)
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
