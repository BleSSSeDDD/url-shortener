package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/BleSSSeDDD/url-shortener/internal/config"
	"github.com/BleSSSeDDD/url-shortener/internal/database"
	"github.com/BleSSSeDDD/url-shortener/internal/handlers"
	"github.com/BleSSSeDDD/url-shortener/internal/service"
	"github.com/BleSSSeDDD/url-shortener/internal/storage"
)

func main() {
	gracefullShutdownCtx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdown()

	serverError := make(chan error, 1) // канал для ошибок сервера

	postgresString := config.GetConnectionStringPostgres()
	redisString := config.GetConnectionStringRedis()

	sqldb, err := database.Init(postgresString)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	defer func() {
		if closeErr := sqldb.Close(); err != nil {
			log.Printf("error closing DB: %v", closeErr)
		}
	}()

	redisdb, err := database.CacheInit(redisString)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	defer func() {
		if err := redisdb.Close(); err != nil {
			log.Printf("error closing Redis: %v", err)
		}
	}()

	rdb, db := storage.NewCache(redisdb), storage.NewPostgres(sqldb)

	shortener := service.NewURLShortener(rdb, db)
	shortenerServer := handlers.NewShortenerServer(shortener)

	go func() {
		if err := shortenerServer.Start(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	log.Println("Сервер запущен")

	//Сценарии конца программы
	select {
	case <-gracefullShutdownCtx.Done():
		log.Println("Сервер будет остановлен по сигналу")
	case err := <-serverError:
		log.Printf("Ошибка сервера: %v\n", err)
		return
	}

	shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := shortenerServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка выключения сервера: %v\n", err)
	}

	log.Println("Сервер остановлен")
}
