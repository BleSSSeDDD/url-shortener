package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	code, err := s.shortener.Set(url)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	shortURL := fmt.Sprintf("http://%s/r/%s", host, code)

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

func (s *ShortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

func (s *ShortenerServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *ShortenerServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

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

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", s.defaultHandler)
	http.HandleFunc("/shorten", s.shortenHandler)
	http.HandleFunc("/r/", s.redirectHandler)
	http.HandleFunc("/health", s.healthHandler)

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

	defer db.Close()

	log.Println("Database reaby")

	shortenerServer := ShortenerServer{shortener: service.NewUrlShortener(db)}

	go func() {
		if err := shortenerServer.Start(); err != nil {
			serverError <- err
		}
		fmt.Println("Сервер запущен")
	}()

	//Сценарии конца программы
	select {
	case <-stop:
		fmt.Println("Сервер остановлен по сигналу")
	case err := <-serverError:
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}
