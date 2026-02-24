package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres interface {
	GetUrlFromCode(code string) (originalUrl string, err error)
	SetNewPair(url string, code string) (string, error)
	Close()
}

func NewPostgres(db *sql.DB) Postgres {
	return &postgres{postgres: db}
}

type postgres struct {
	postgres *sql.DB
}

func (db *postgres) GetUrlFromCode(code string) (originalUrl string, err error) {
	row := db.postgres.QueryRow("SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalUrl)
	return originalUrl, err
}

// если все норм, вернет nil
func (db *postgres) SetNewPair(url string, code string) (string, error) {
	var resultCode string
	err := db.postgres.QueryRow(
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

func (db *postgres) Close() {
	db.postgres.Close()
}
