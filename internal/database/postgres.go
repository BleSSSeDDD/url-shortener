// Package database нужен для подключения к бдшке и кэшу
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PostgresPingTimeout — таймаут проверки соединения с postgres при старте.
const PostgresPingTimeout = 5 * time.Second

// Init создает новое подключение к базе данных и проверяет его пингом.
func Init(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("инициализация подключения к postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), PostgresPingTimeout)
	defer cancel()

	// sql.Open не открывает соединение, реальную доступность проверяем пингом.
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("ошибка закрытия соединения после неудачного ping: %v", closeErr)
		}
		return nil, fmt.Errorf("подключение к postgres: %w", err)
	}

	log.Println("Подключились к постгресу")
	return db, nil
}
