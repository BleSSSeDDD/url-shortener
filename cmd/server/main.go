package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/BleSSSeDDD/url-shortener/internal/app"
)

// ShortenerServer нужен чтобы инкапсулировать UrlShortener с методами самого сервера, которые отношенеия к внутренней логике вообще не имеют
type ShortenerServer struct {
	shortener *app.UrlShortener
}

// Хендлер для /shorten
func (s *ShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")

}

// Хендлер для корневой директории сервера
func (s *ShortenerServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		w.Write([]byte("Ошибка сервера, html не прочитался"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

// Стартует сервер на порту 8080, если порт занят или другая ошибка - возвращает её
func (s *ShortenerServer) Start() error {
	fmt.Println("Запускаем сервер")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", s.defaultHandler)
	http.HandleFunc("/shorten", s.shortenHandler)

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		return err
	}

	return nil
}

func main() {
	stop := make(chan os.Signal, 1) //для грейсфул шатдауна
	signal.Notify(stop, os.Interrupt)

	serverError := make(chan error, 1) // канал для ошибок сервера

	shortenerServer := ShortenerServer{shortener: app.NewUrlShortener()}

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
		os.Exit(1)
	}
}
