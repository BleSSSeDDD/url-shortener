// Package storage для работы с базой данных
package storage

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq" // бланк импорт потому что драйверы так импортируются
)

// Postgres интерфейс для мока бдшки
type Postgres interface {
	GetURLFromCode(ctx context.Context, code string) (originalURL string, err error)
	SetNewPair(ctx context.Context, url string, code string) (string, error)
}

// NewPostgres создает экземпляр структуры postgres для работы с бдшкой
func NewPostgres(db *sql.DB) Postgres {
	return &postgres{postgres: db}
}

type postgres struct {
	postgres *sql.DB
}

// GetUrlFromCode идёт в базу данных и ищет там ссылку, на которую ссылается код
func (db *postgres) GetURLFromCode(ctx context.Context, code string) (originalURL string, err error) {
	row := db.postgres.QueryRowContext(ctx, "SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalURL)
	return originalURL, err
}

// если все норм, вернет nil
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
