package app

import (
	"errors"
	"fmt"
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
	fmt.Println("Создан новый UrlShortener")
	return &UrlShortener{
		codeToURL: make(map[string]string),
		urlToCode: make(map[string]string),
		mutex:     sync.RWMutex{},
	}
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
	fmt.Printf("🔧 Set() вызван с URL: %s\n", url)
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Проверяем существующий URL
	if existingCode, exists := u.urlToCode[url]; exists {
		fmt.Printf("URL уже существует, возвращаем существующий код: %s\n", existingCode)
		return existingCode, nil
	}

	fmt.Printf("Создаём новую короткую ссылку для: %s\n", url)

	// Генерируем новый уникальный код
	for i := 0; i < 10; i++ {
		code := generateShortenedUrl()
		fmt.Printf("Попытка %d: сгенерирован код %s\n", i+1, code)

		if _, exists := u.codeToURL[code]; !exists {
			u.codeToURL[code] = url
			u.urlToCode[url] = code
			fmt.Printf("Успешно! Сохранено: %s -> %s\n", code, url)
			fmt.Printf("Текущее количество ссылок: %d\n", len(u.codeToURL))
			return code, nil
		} else {
			fmt.Printf(" Код %s уже существует, пробуем другой\n", code)
		}
	}

	fmt.Printf("Не удалось сгенерировать уникальный код за 10 попыток\n")
	return "", errors.New("не удалось сгенерировать уникальный код")
}

// Если ссылка есть, мы отдаем её, если нет то пустую строку и ошибку
func (u *UrlShortener) Get(shortCode string) (originalUrl string, err error) {
	fmt.Printf("Get() вызван с кодом: %s\n", shortCode)

	u.mutex.RLock()
	defer u.mutex.RUnlock()

	originalUrl, exists := u.codeToURL[shortCode]
	if !exists {
		fmt.Printf("Код %s не найден в базе\n", shortCode)
		fmt.Printf("Доступные коды: %v\n", u.getAvailableCodes())
		return "", errors.New("нет такой ссылки")
	}

	fmt.Printf("Найден URL для кода %s: %s\n", shortCode, originalUrl)
	return originalUrl, nil
}

// Вспомогательный метод для логирования доступных кодов
func (u *UrlShortener) getAvailableCodes() []string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	codes := make([]string, 0, len(u.codeToURL))
	for code := range u.codeToURL {
		codes = append(codes, code)
	}
	return codes
}
