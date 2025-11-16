package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"text/template"

	"github.com/BleSSSeDDD/url-shortener/internal/app"
)

// ShortenerServer нужен чтобы инкапсулировать UrlShortener с методами самого сервера, которые отношенеия к внутренней логике вообще не имеют
type ShortenerServer struct {
	shortener *app.UrlShortener
}

// Хендлер для /shorten
func (s *ShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("📨 Получен запрос на /shorten методом %s\n", r.Method)

	if r.Method != http.MethodPost {
		fmt.Printf("❌ Неподдерживаемый метод: %s\n", r.Method)
		http.Error(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	fmt.Printf("🔗 Получен URL для сокращения: %s\n", url)

	if url == "" {
		fmt.Println("❌ Пустой URL")
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	fmt.Println("🔄 Вызываем shortener.Set...")
	shortCode, err := s.shortener.Set(url)
	if err != nil {
		fmt.Printf("❌ Ошибка в shortener.Set: %v\n", err)
		w.Write([]byte("Ошибка на стороне сервера, попробуйте снова :("))
		return
	}

	shortURL := "http://" + r.Host + "/" + shortCode
	fmt.Printf("✅ Успешно создана короткая ссылка: %s -> %s\n", url, shortCode)

	// Парсим HTML шаблон
	tmpl, err := template.ParseFiles("./templates/shorten.html")
	if err != nil {
		fmt.Printf("❌ Ошибка парсинга шаблона: %v\n", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	data := struct {
		ShortURL    string
		OriginalURL string
	}{
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	// Рендерим шаблон
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Printf("❌ Ошибка рендеринга шаблона: %v\n", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	fmt.Println("✅ HTML результат отправлен")
}

// Дефолт хендлер либо отдает / либо если это что-то другое, то редиректит с redirectHandler
func (s *ShortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🌐 Запрос к defaultHandler: %s\n", r.URL.Path)

	if r.URL.Path != "/" {
		fmt.Printf("🔄 Перенаправляем в redirectHandler: %s\n", r.URL.Path)
		s.redirectHandler(w, r)
		return
	}

	fmt.Println("📄 Отдаём HTML форму")
	htmlContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		fmt.Printf("❌ Ошибка чтения HTML: %v\n", err)
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
	fmt.Println("✅ HTML форма успешно отправлена")
}

// Ко всему кроме / и /shorten относимся как к сокращенной ссылке
func (s *ShortenerServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]
	fmt.Printf("🔍 Поиск кода для редиректа: '%s'\n", shortCode)

	originalURL, err := s.shortener.Get(shortCode)
	if err != nil {
		fmt.Printf("❌ Код '%s' не найден: %v\n", shortCode, err)
		http.NotFound(w, r)
		return
	}

	fmt.Printf("🔄 Редирект с %s на %s\n", shortCode, originalURL)
	http.Redirect(w, r, originalURL, http.StatusFound)
	fmt.Printf("✅ Редирект выполнен\n")
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *ShortenerServer) Start() error {
	fmt.Println("🚀 Запускаем сервер на localhost:8080")
	fmt.Println("📊 Маршруты:")
	fmt.Println("   GET  /       - форма для ввода URL")
	fmt.Println("   POST /shorten - создание короткой ссылки")
	fmt.Println("   GET  /<code> - редирект по короткой ссылке")

	http.HandleFunc("/", s.defaultHandler)
	http.HandleFunc("/shorten", s.shortenHandler)

	fmt.Println("✅ Сервер запущен и слушает порт 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("💥 Ошибка сервера: %v\n", err)
		return err
	}

	return nil
}

func main() {
	fmt.Println("🔧 Инициализация URL Shortener...")

	stop := make(chan os.Signal, 1) //для грейсфул шатдауна
	signal.Notify(stop, os.Interrupt)

	serverError := make(chan error, 1) // канал для ошибок сервера

	shortenerServer := ShortenerServer{shortener: app.NewUrlShortener()}
	fmt.Println("✅ UrlShortener создан")

	go func() {
		fmt.Println("🔄 Запуск сервера в горутине...")
		if err := shortenerServer.Start(); err != nil {
			fmt.Printf("💥 Ошибка при запуске сервера: %v\n", err)
			serverError <- err
		}
	}()

	fmt.Println("⏳ Сервер запущен. Ожидание запросов...")
	fmt.Println("🛑 Для остановки нажмите Ctrl+C")

	//Сценарии конца программы
	select {
	case <-stop:
		fmt.Println("\n🛑 Получен сигнал остановки")
		fmt.Println("👋 Завершение работы...")
	case err := <-serverError:
		fmt.Printf("💥 Критическая ошибка сервера: %v\n", err)
		os.Exit(1)
	}
}
