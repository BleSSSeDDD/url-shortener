package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BleSSSeDDD/url-shortener/internal/config"
	"github.com/BleSSSeDDD/url-shortener/internal/database"
	"github.com/BleSSSeDDD/url-shortener/internal/handlers"
	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

func main() {
	stop := make(chan os.Signal, 1) //для грейсфул шатдауна
	signal.Notify(stop, syscall.SIGTERM)

	serverError := make(chan error, 1) // канал для ошибок сервера

	postgresString := config.GetConnectionStringPostgres()
	redisString := config.GetConnectionStringRedis()

	redisdb, err := database.CacheInit(redisString)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	sqldb, err := database.Init(postgresString)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	defer func() {
		if err := sqldb.Close(); err != nil {
			log.Printf("error closing DB: %v", err)
		}
	}()

	rdb, db := storage.NewCache(redisdb), storage.NewPostgres(sqldb)

	shortenerServer := service.NewUrlShortener(rdb, db)
	handlerInterface := handlers.NewShortenerServer(shortenerServer)

	go func() {
		if err := handlerInterface.Start(); err != nil {
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
