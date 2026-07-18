// Package storage для работы с базой данных
package storage

import (
	"context"
	"database/sql"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	_ "github.com/lib/pq" // пустой импорт: регистрирует драйвер postgres в database/sql
)

// NewStorage создаёт реализацию доступа к postgres (чтение и запись пар url/code)
func NewStorage(db *sql.DB) (service.StorageGetter, service.StorageSetter) {
	pg := &postgres{postgres: db}
	return pg, pg
}

type postgres struct {
	postgres *sql.DB
}

// GetURLFromCode идёт в базу данных и ищет там ссылку, на которую ссылается код
func (db *postgres) GetURLFromCode(ctx context.Context, code string) (originalURL string, err error) {
	row := db.postgres.QueryRowContext(ctx, "SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalURL)
	return originalURL, err
}

// SetNewPair вставляет пару url/code. Если url уже есть, ON CONFLICT (url)
// возвращает существующий код без ошибки.
func (db *postgres) SetNewPair(ctx context.Context, url string, code string) (string, error) {
	var resultCode string
	err := db.postgres.QueryRowContext(ctx,
		`INSERT INTO urls_and_codes (url, code) VALUES ($1, $2)
         ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
         RETURNING code`,
		url, code,
	).Scan(&resultCode)

	if err != nil {
		return "", err
	}
	return resultCode, nil
}
