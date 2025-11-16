package app

import (
	"errors"
	"math/rand"
	"sync"
)

// Содержит поля с ин-мемори хранилищем и мьютексом, методами можно сохранять новые значения или получать старые
type UrlShortener struct {
	urls  map[string]string
	mutex sync.RWMutex
}

// Создает структуру UrlShortener, возвращает на неё указатель
func NewUrlShortener() *UrlShortener {
	return &UrlShortener{urls: make(map[string]string), mutex: sync.RWMutex{}}
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

	encodedUrl := generateShortenedUrl()

	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()
		if _, exists := u.urls[code]; !exists {
			u.urls[code] = url
			return code, nil
		}
	}

	return encodedUrl, nil
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

	res, exists := u.urls[encodedUrl]

	if !exists {
		return "", errors.New("url не найден")
	}

	return res, nil
}
