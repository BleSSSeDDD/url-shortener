package storage

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres interface {
	GetUrlFromCode(ctx context.Context, code string) (originalUrl string, err error)
	SetNewPair(ctx context.Context, url string, code string) (string, error)
}

func NewPostgres(db *sql.DB) Postgres {
	return &postgres{postgres: db}
}

type postgres struct {
	postgres *sql.DB
}

func (db *postgres) GetUrlFromCode(ctx context.Context, code string) (originalUrl string, err error) {
	row := db.postgres.QueryRowContext(ctx, "SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalUrl)
	return originalUrl, err
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
