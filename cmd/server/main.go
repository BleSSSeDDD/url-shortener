package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/handlers"
	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

func main() {
	stop := make(chan os.Signal, 1) //для грейсфул шатдауна
	signal.Notify(stop, syscall.SIGTERM)

	serverError := make(chan error, 1) // канал для ошибок сервера

	rdb, err := storage.CacheInit()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	var db *sql.DB
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

	shortenerServer := handlers.ShortenerServer{Shortener: service.NewUrlShortener(db, rdb)}

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
