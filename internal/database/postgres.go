package database

import (
	"database/sql"
	"log"
)

func Init(connectionString string) (db *sql.DB, err error) {
	dbAttempt, err1 := sql.Open("postgres", connectionString)
	if err1 != nil {
		log.Println("Неудачная попытка подключиться к бд, пробуем...")
		return nil, err
	}
	log.Println("Подключились к постгресу")
	return dbAttempt, nil
}
