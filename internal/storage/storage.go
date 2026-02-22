package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres interface {
	GetUrlFromCode(code string) (originalUrl string, err error)
	GetCodeFromUrl(url string) (code string, err error)
	SetNewPair(url string, code string) error
	Close()
}

func NewPostgres(db *sql.DB) Postgres {
	return &postgres{postgres: db}
}

type postgres struct {
	postgres *sql.DB
}

func InitializeDB(db *sql.DB) Postgres {
	return &postgres{postgres: db}
}

func (db *postgres) GetUrlFromCode(code string) (originalUrl string, err error) {
	row := db.postgres.QueryRow("SELECT url FROM urls_and_codes WHERE code = $1", code)
	err = row.Scan(&originalUrl)
	return originalUrl, err
}

func (db *postgres) GetCodeFromUrl(url string) (code string, err error) {
	row := db.postgres.QueryRow("SELECT code FROM urls_and_codes WHERE url = $1", url)
	err = row.Scan(&code)
	return code, err
}

// если все норм, вернет nil
func (db *postgres) SetNewPair(url string, code string) error {
	_, err := db.postgres.Exec("INSERT INTO urls_and_codes (url, code) VALUES ($1, $2)", url, code)
	return err
}

func (db *postgres) Close() {
	db.postgres.Close()
}
