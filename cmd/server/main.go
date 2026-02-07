package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

type ShortenerServer struct {
	shortener *service.UrlShortener
}

func (s *ShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")

	shortenedUrl, err := s.shortener.Set(url)
	if err != nil {
		w.Write([]byte("Ошибка на стороне сервера, попробуйте снова :("))
		fmt.Println(err)
		return
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	w.Write([]byte("http://" + host + "/" + shortenedUrl))
}

func (s *ShortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.redirectHandler(w, r)
		return
	}

	htmlContent, err := os.ReadFile("/server/templates/index.html")
	if err != nil {
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

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

func (s *ShortenerServer) Start(port int) error {
	fmt.Printf("Запускаем сервер на порту %d\n", port)

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
		if err := shortenerServer.Start(port); err != nil {
			serverError <- err
		}
	}()

	fmt.Printf("Сервер запущен и доступен по адресу http://0.0.0.0:%d\n", port)

	select {
	case <-stop:
		fmt.Println("Сервер остановлен по сигналу")
	case err := <-serverError:
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}
