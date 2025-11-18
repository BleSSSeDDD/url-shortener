package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// Инициализирует бд, возвращает указатель на неё и ошибку
func Init(dbPath string) (*sql.DB, error) {
	psqlConfig, err := os.ReadFile("test.env")
	if err != nil {
		return nil, fmt.Errorf("енв файл не прочитался: %v", err)
	}

	db, err := sql.Open("postgres", string(psqlConfig))

	if err != nil {
		return nil, err
	}

	if !tablesExist(db) {
		log.Println("База данных не найдена, надо создать новую")
		if err := createTables(db); err != nil {
			return nil, err
		}
		log.Println("Таблицы успешно созданы, бд подключена")
	} else {
		log.Println("База данных подключена")
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Пинг не прошёл:", err)
		return nil, err
	}

	return db, nil
}

// Проверяет, существую ли бд и возвращает true или false
func tablesExist(db *sql.DB) bool {
	tables := []string{"urls_and_codes"}

	for _, table := range tables {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		err := db.QueryRow(query, table).Scan(&name)
		if err != nil || name != table {
			return false
		}
	}
	return true
}

// Если таблицы еще нет  в бд, то создает её
func createTables(db *sql.DB) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS urls_and_codes (
			url INTEGER,                      
			url-code INTEGER
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// Делает запрос в бд и возвращает оригинальную ссылку и ошибку
func GetUrl(db *sql.DB, code string) {

}

// Делает запрос в бд и возвращает короткий код и ошибку
func GetCode(db *sql.DB, url string) {

}

// создает новую запись в бд с оригинальной ссылкой
func NewUrl(db *sql.DB, newUrl string) {

}

// Создает новую запись в бд с кодом
func NewCode(db *sql.DB, newCode string) {

}

//что-то из этого не понадобится
