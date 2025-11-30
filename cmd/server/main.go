package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/BleSSSeDDD/url-shortener/internal/app"
)

type ShortenerServer struct {
	shortener *app.UrlShortener
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

	htmlContent, err := os.ReadFile("./templates/index.html")
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

	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return err
	}

	return nil
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	serverError := make(chan error, 1)

	shortenerServer := ShortenerServer{shortener: app.NewUrlShortener()}

	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

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
		os.Exit(1)
	}
}
