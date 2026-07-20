// Package handlers пакет для работы с хттп
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// maxURLLength соответствует VARCHAR(500) в схеме БД.
const maxURLLength = 500

// validateURL проверяет, что строка — корректный http(s) URL допустимой длины.
func validateURL(raw string) error {
	if len(raw) > maxURLLength {
		return fmt.Errorf("URL длиннее %d символов", maxURLLength)
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("некорректный URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("недопустимая схема %q: разрешены только http и https", parsed.Scheme)
	}

	if parsed.Host == "" {
		return errors.New("в URL отсутствует хост")
	}

	return nil
}

// URLShortenerGetter интерфейс для бизнес-логики получения ссылок
type URLShortenerGetter interface {
	Get(ctx context.Context, shortCode string) (originalURL string, err error)
}

// URLShortener интерфейс для бизнес-логики получения и записи ссылок
type URLShortenerSetter interface {
	Set(ctx context.Context, url string) (shortenedURL string, err error)
}

// NewShortenerServer нужна для DI, инкапсулирует бизнес-логику с хттп
func NewShortenerServer(shortenerGetter URLShortenerGetter, shortenerSetter URLShortenerSetter) ShortenerServer {
	return &shortenerServer{shortenerGetter: shortenerGetter, shortenerSetter: shortenerSetter}
}

// ShortenerServer интерфейс для того, чтоб можно было мокать сервер
type ShortenerServer interface {
	Start() error
	Shutdown(shutdownCtx context.Context) error
}

// ShortenerServer нужен чтобы инкапсулировать UrlShortener с методами самого сервера, которые отношенеия к внутренней логике вообще не имеют
type shortenerServer struct {
	shortenerGetter URLShortenerGetter
	shortenerSetter URLShortenerSetter
	server          *http.Server
	metricsServer   *http.Server
	templates       *template.Template
}

func (s *shortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	rawURL := r.FormValue("url")
	if rawURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if err := validateURL(rawURL); err != nil {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	code, err := s.shortenerSetter.Set(r.Context(), rawURL)
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

	data := struct {
		OriginalURL string
		ShortURL    string
	}{rawURL, shortURL}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "shorten.html", data); err != nil {
		log.Printf("Ошибка выполнения шаблона shorten.html: %v", err)
	}
}

func (s *shortenerServer) defaultHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Printf("ошибка выполнения шаблона index.html: %v", err)
	}
}

func (s *shortenerServer) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("ошибка записи healthcheck: %v", err)
	}
}

func (s *shortenerServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "code")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	log.Printf("Поиск кода: %s\n", shortCode)

	originalURL, err := s.shortenerGetter.Get(r.Context(), shortCode)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log.Printf("Редирект с %s на %s\n", shortCode, originalURL)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (s *shortenerServer) apiRootHandler(w http.ResponseWriter, _ *http.Request) {
	response := APIResponse{
		Service:       "URL Shortener API",
		Versions:      []string{"v1"},
		Latest:        "v1",
		Documentation: "/api/v1",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
	}
}

func (s *shortenerServer) apiV1RootHandler(w http.ResponseWriter, _ *http.Request) {
	response := APIV1Response{
		Version: "v1",
		Status:  "active",
		Endpoints: []EndpointInfo{
			{Path: "/health", Method: "GET", Description: "Health check"},
			{Path: "/shorten", Method: "POST", Description: "Create short URL"},
		},
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
	}
}

func (s *shortenerServer) healthAPIHandler(w http.ResponseWriter, _ *http.Request) {
	response := HealthResponse{
		Status:  "ok",
		Service: "url-shortener",
		Version: "v1",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
	}
}

func (s *shortenerServer) shortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"only POST allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, `{"error":"url required"}`, http.StatusBadRequest)
		return
	}

	if err := validateURL(req.URL); err != nil {
		http.Error(w, `{"error":"invalid url"}`, http.StatusBadRequest)
		return
	}

	code, err := s.shortenerSetter.Set(r.Context(), req.URL)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("http://%s/r/%s", r.Host, code)

	response := ShortenResponse{
		ShortURL: shortURL,
		Code:     code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
	}
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *shortenerServer) Start() error {
	tmpl, err := template.ParseFiles("./templates/index.html", "./templates/shorten.html")
	if err != nil {
		return fmt.Errorf("парсинг HTML-шаблонов: %w", err)
	}
	s.templates = tmpl

	r := chi.NewRouter()

	r.Use(Metrics)

	// статика
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// api general
	r.Get("/api", s.apiRootHandler)

	// api v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", s.apiV1RootHandler)
		r.Get("/health", s.healthAPIHandler)
		r.Post("/shorten", s.shortenAPIHandler)
	})

	// html
	r.HandleFunc("/", s.defaultHandler)
	r.HandleFunc("/shorten", s.shortenHandler)
	r.HandleFunc("/r/{code}", s.redirectHandler)
	r.Get("/health", s.healthHandler)

	s.server = &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// Отдельный сервер только для метрик на внутреннем порту 9100. Наружу этот
	// порт не публикуется — /metrics доступен лишь Prometheus по внутренней
	// сети, поэтому прятать его на уровне traefik больше не нужно.
	s.metricsServer = &http.Server{
		Addr:              ":9100",
		Handler:           promhttp.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := s.metricsServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("сервер метрик остановлен с ошибкой: %v", err)
		}
	}()

	return s.server.ListenAndServe()
}

func (s *shortenerServer) Shutdown(shutdownCtx context.Context) error {
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("ошибка остановки сервера метрик: %v", err)
		}
	}

	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(shutdownCtx)
}
