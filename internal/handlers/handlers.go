package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
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
		log.Printf("Ошибка парсинга шаблона shorten.html: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
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
		log.Printf("Ошибка парсинга шаблона index.html: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
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
	shortCode := chi.URLParam(r, "code")
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

func (s *shortenerServer) apiRootHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Service:       "URL Shortener API",
		Versions:      []string{"v1"},
		Latest:        "v1",
		Documentation: "/api/v1",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *shortenerServer) apiV1RootHandler(w http.ResponseWriter, r *http.Request) {
	response := APIV1Response{
		Version: "v1",
		Status:  "active",
		Endpoints: []EndpointInfo{
			{Path: "/health", Method: "GET", Description: "Health check"},
			{Path: "/shorten", Method: "POST", Description: "Create short URL"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *shortenerServer) healthAPIHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ok",
		Service: "url-shortener",
		Version: "v1",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	code, err := s.shortener.Set(req.URL)
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
	json.NewEncoder(w).Encode(response)
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *shortenerServer) Start() error {
	r := chi.NewRouter()

	//статика
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/", http.StripPrefix("/static/", fileServer))

	//api general
	r.Get("/api", s.apiRootHandler)

	//api v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", s.apiV1RootHandler)
		r.Get("/health", s.healthAPIHandler)
		r.Post("/shorten", s.shortenAPIHandler)
	})

	//html
	r.HandleFunc("/", s.defaultHandler)
	r.HandleFunc("/shorten", s.shortenHandler)
	r.HandleFunc("/r/{code}", s.redirectHandler)
	r.Get("/health", s.healthHandler)

	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}

	return nil
}
