// Package storage provides database access.
package storage

import (
	"context"
	"database/sql"

	"github.com/BleSSSeDDD/url-shortener/internal/service"
	_ "github.com/lib/pq" // blank import: registers the postgres driver with database/sql
)

// NewStorage builds the postgres-backed store for url/code pairs.
func NewStorage(db *sql.DB) (service.StorageGetter, service.StorageSetter) {
	pg := &postgres{postgres: db}
	return pg, pg
}

type postgres struct {
	postgres *sql.DB
}

// GetURLFromCode looks up the URL a short code points to.
func (db *postgres) GetURLFromCode(ctx context.Context, code string) (originalURL string, err error) {
	row := db.postgres.QueryRowContext(ctx, "SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalURL)
	return originalURL, err
}

// SetNewPair inserts a url/code pair. If the url already exists, ON CONFLICT (url)
// returns the existing code without an error.
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
