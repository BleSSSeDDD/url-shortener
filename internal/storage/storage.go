package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Postgres interface {
	GetUrlFromCode(code string) (originalUrl string, err error)
	GetCodeFromUrl(url string) (code string, err error)
	SetNewPair(url string, code string) error
	Close()
}

type postgres struct {
	postgres *sql.DB
}

func Init() (Postgres, error) {
	connectString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s  sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &postgres{postgres: db}, nil
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
