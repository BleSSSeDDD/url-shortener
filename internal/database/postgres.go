// Package database нужен для подключения к бдшке и кэшу
package database

import (
	"database/sql"
	"log"
)

// Init создает новое подключение к базе данных
func Init(connectionString string) (db *sql.DB, err error) {
	dbAttempt, err1 := sql.Open("postgres", connectionString)
	if err1 != nil {
		log.Println("Неудачная попытка подключиться к бд, пробуем...")
		return nil, err1
	}
	log.Println("Подключились к постгресу")
	return dbAttempt, nil
}
