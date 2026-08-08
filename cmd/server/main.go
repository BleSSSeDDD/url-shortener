// Package main starts the HTTP server with all of its dependencies.
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
		// Non-zero exit code so the orchestrator can tell the start failed.
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	gracefulCtx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdown()

	serverError := make(chan error, 1) // carries errors raised by the HTTP server

	postgresString := config.GetConnectionStringPostgres()
	redisString := config.GetConnectionStringRedis()

	sqldb, err := database.Init(postgresString)
	if err != nil {
		return fmt.Errorf("init postgres: %w", err)
	}

	defer func() {
		if closeErr := sqldb.Close(); closeErr != nil {
			log.Printf("failed to close postgres: %v", closeErr)
		}
	}()

	redisdb, err := database.CacheInit(redisString)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}

	defer func() {
		if closeErr := redisdb.Close(); closeErr != nil {
			log.Printf("failed to close redis: %v", closeErr)
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

	log.Println("server started")

	// Shutdown paths
	select {
	case <-gracefulCtx.Done():
		log.Println("server is shutting down on signal")
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)
	}

	shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := shortenerServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Println("server stopped")
	return nil
}
