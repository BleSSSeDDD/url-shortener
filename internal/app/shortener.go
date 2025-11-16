package app

import (
	"errors"
	"math/rand"
	"sync"
)

// Содержит поля с ин-мемори хранилищем и мьютексом, методами можно сохранять новые значения или получать старые
type UrlShortener struct {
	codeToURL map[string]string // код -> URL
	urlToCode map[string]string // URL -> код
	mutex     sync.RWMutex
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener() *UrlShortener {
	return &UrlShortener{codeToURL: make(map[string]string), urlToCode: make(map[string]string), mutex: sync.RWMutex{}}
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
func (u *UrlShortener) Set(url string) (shortenedUrl string, err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if existingCode, exists := u.urlToCode[url]; exists {
		return existingCode, nil
	}

	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()
		if _, exists := u.codeToURL[code]; !exists {
			u.codeToURL[code] = url
			u.urlToCode[url] = code
			return code, nil
		}
	}

	return "", errors.New("не удалось сгенерировать уникальный код")
}

// Принимает: сокращенный код
//
// Возвращает: оригинальную ссылку и ошибку
//
// Логика: ищет код в базе, возвращает оригинальный url
// либо если его нет то пустую строку и ошибку
func (u *UrlShortener) Get(encodedUrl string) (shortenedUrl string, err error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	res, exists := u.codeToURL[encodedUrl]

	if !exists {
		return "", errors.New("url не найден")
	}

	return res, nil
}
