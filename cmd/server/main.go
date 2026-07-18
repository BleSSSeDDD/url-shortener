// основной пакет для запуска сервера со всеми зависимостями
package main

import (
	"context"
	"errors"
	"fmt"
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
	if err := run(); err != nil {
		// Ненулевой код выхода, чтобы оркестратор понял, что старт провалился.
		log.Fatalf("фатальная ошибка: %v", err)
	}
}

func run() error {
	gracefulCtx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdown()

	serverError := make(chan error, 1) // канал для ошибок сервера

	postgresString := config.GetConnectionStringPostgres()
	redisString := config.GetConnectionStringRedis()

	sqldb, err := database.Init(postgresString)
	if err != nil {
		return fmt.Errorf("инициализация postgres: %w", err)
	}

	defer func() {
		if closeErr := sqldb.Close(); closeErr != nil {
			log.Printf("ошибка закрытия postgres: %v", closeErr)
		}
	}()

	redisdb, err := database.CacheInit(redisString)
	if err != nil {
		return fmt.Errorf("инициализация redis: %w", err)
	}

	defer func() {
		if closeErr := redisdb.Close(); closeErr != nil {
			log.Printf("ошибка закрытия redis: %v", closeErr)
		}
	}()

	cacheGetter, cacheSetter := storage.NewCache(redisdb)
	storageGetter, storageSetter := storage.NewStorage(sqldb)

	shortenerGetter, shortenerSetter := service.NewURLShortener(cacheGetter, cacheSetter, storageGetter, storageSetter)
	shortenerServer := handlers.NewShortenerServer(shortenerGetter, shortenerSetter)

	go func() {
		if err := shortenerServer.Start(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	log.Println("Сервер запущен")

	// Сценарии конца программы
	select {
	case <-gracefulCtx.Done():
		log.Println("Сервер будет остановлен по сигналу")
	case err := <-serverError:
		return fmt.Errorf("ошибка сервера: %w", err)
	}

	shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := shortenerServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("выключение сервера: %w", err)
	}

	log.Println("Сервер остановлен")
	return nil
}
