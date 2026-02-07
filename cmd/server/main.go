package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

// ShortenerServer нужен чтобы инкапсулировать UrlShortener с методами самого сервера, которые отношенеия к внутренней логике вообще не имеют
type ShortenerServer struct {
	shortener *service.UrlShortener
}

func (s *ShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", 400)
		return
	}

	code, err := s.shortener.Set(url)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Server error", 500)
		return
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	shortURL := fmt.Sprintf("http://%s/%s", host, code)

	tmpl, err := template.ParseFiles("./templates/shorten.html")
	if err != nil {
		fmt.Fprintf(w, "<br>Сокращённая: %s<br><a href='/'>Назад</a>", shortURL)
		return
	}

	data := struct {
		OriginalURL string
		ShortURL    string
	}{url, shortURL}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// Дефолт хендлер либо отдает / либо стайл.css если он запрашивается, если это что-то другое, то редиректит с redirectHandler
func (s *ShortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/static/style.css" {
		http.ServeFile(w, r, "./static/style.css")
		return
	}

	if r.URL.Path != "/" {
		s.redirectHandler(w, r)
		return
	}

	htmlContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

// Ко всему кроме / и /shorten относимся как к сокращенной ссылке
func (s *ShortenerServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]

	fmt.Printf("Поиск кода: %s\n", shortCode)

	originalURL, err := s.shortener.Get(shortCode)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Printf("Редирект с %s на %s\n", shortCode, originalURL)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *ShortenerServer) Start() error {
	fmt.Println("Запускаем сервер")

	http.HandleFunc("/", s.defaultHandler)
	http.HandleFunc("/shorten", s.shortenHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}

	return nil
}

func main() {
	stop := make(chan os.Signal, 1) //для грейсфул шатдауна
	signal.Notify(stop, syscall.SIGTERM)

	serverError := make(chan error, 1) // канал для ошибок сервера

	var db *sql.DB
	var err error
	for range 5 {
		db, err = storage.Init()
		if err != nil {
			log.Printf("Error: %v, retrying...\n", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	if err != nil {
		log.Printf("Error: %v, could not connect to database\n", err)
		return
	}

	log.Println("Database reaby")

	shortenerServer := ShortenerServer{shortener: service.NewUrlShortener(db)}

	go func() {
		if err := shortenerServer.Start(); err != nil {
			serverError <- err
		}
	}()

	//Сценарии конца программы
	select {
	case <-stop:
		fmt.Println("Сервер остановлен по сигналу")
	case err := <-serverError:
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}
