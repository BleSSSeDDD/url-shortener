package service

import (
	"database/sql"
	"math/rand"

	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

type UrlShortener struct {
	db *sql.DB
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener(db *sql.DB) *UrlShortener {
	return &UrlShortener{db: db}
}

// Генерирует случайную строку из 6 символов
func generateShortenedUrl() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var res string

	for i := 0; i < 6; i++ {
		res += string(charset[rand.Intn(len(charset))])
	}

	return res
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

	code := generateShortenedUrl()

	// Генерируем новый уникальный код
	for err == nil {
		code = generateShortenedUrl()
		_, err = storage.GetUrlFromCode(u.db, code)
	}

	err = nil

	if seterr := storage.SetNewPair(u.db, url, code); seterr != nil {
		return "", seterr
	}

	return code, nil
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *UrlShortener) Get(shortCode string) (originalUrl string, err error) {

	originalUrl, err = storage.GetUrlFromCode(u.db, shortCode)

	return originalUrl, err
}
