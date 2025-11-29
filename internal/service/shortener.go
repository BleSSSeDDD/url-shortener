package service

import (
	"errors"
	"math/rand"
)

// Содержит два поля, представляющие сущности из бд: ссылку и её короткий код.
type UrlShortener struct {
	originalURL string
	encodedUrl  string
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener() *UrlShortener {
	return &UrlShortener{}
}

// Генерирует случайную строку из 4 символов, коллизия на 10000 генераций 0,3%
func generateShortenedUrl() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var res string

	for i := 0; i < 4; i++ {
		res += string(charset[rand.Intn(len(charset))])
	}

	return res
}

// Принимает: оригинальный URL
//
// Возвращает: короткий код и ошибку
//
// Логика: генерирует код до 10 раз чтобы избежать коллизии, сохраняет в базу, возвращает код
//
// ЛИБО если такое уже есть, то отдаём чё есть
func (u *UrlShortener) Set(url string) (shortenedUrl string, err error) {

	// Проверяем существующий URL
	if existingCode, exists := u.urlToCode[url]; exists {
		return existingCode, nil
	}

	// Генерируем новый уникальный код
	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()

		// if _, exists := u.codeToURL[code]; !exists {
		// 	u.codeToURL[code] = url
		// 	u.urlToCode[url] = code
		// 	return code, nil
		// }
	}

	return "", errors.New("не удалось сгенерировать уникальный код")
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *UrlShortener) Get(shortCode string) (originalUrl string, err error) {

	originalUrl, exists := u.codeToURL[shortCode]
	if !exists {
		return "", errors.New("нет такой ссылки")
	}

	return originalUrl, nil
}
