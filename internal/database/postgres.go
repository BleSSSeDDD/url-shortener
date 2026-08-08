// Package database opens connections to the database and the cache.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PostgresPingTimeout bounds the startup connectivity check against postgres.
const PostgresPingTimeout = 5 * time.Second

// Init opens a new database connection and verifies it with a ping.
func Init(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), PostgresPingTimeout)
	defer cancel()

	// sql.Open does not open a connection, so reachability is verified with a ping.
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close connection after a failed ping: %v", closeErr)
		}
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	log.Println("connected to postgres")
	return db, nil
}
