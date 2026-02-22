package config

import (
	"fmt"
	"os"
)

func getEnv(key string, defaulValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaulValue
}

// проверки для Getenv что она непустая
func GetConnectionStringPostgres() string {
	connectString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "postgres"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "12345678"),
		getEnv("DB_NAME", "urls_and_codes"),
	)

	return connectString
}

func GetConnectionStringRedis() string {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	return addr
}
