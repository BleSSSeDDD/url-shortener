package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
)

func NewShortenerServer(shortener service.UrlShortener) ShortenerServer {
	return &shortenerServer{shortener: shortener}
}

type ShortenerServer interface {
	Start() error
}

// ShortenerServer нужен чтобы инкапсулировать UrlShortener с методами самого сервера, которые отношенеия к внутренней логике вообще не имеют
type shortenerServer struct {
	shortener service.UrlShortener
}

func (s *shortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *shortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

func (s *shortenerServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *shortenerServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	log.Printf("Поиск кода: %s\n", shortCode)

	originalURL, err := s.shortener.Get(shortCode)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log.Printf("Редирект с %s на %s\n", shortCode, originalURL)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *shortenerServer) Start() error {

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
