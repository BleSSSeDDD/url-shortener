package config

import (
	"fmt"
	"os"
)

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
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
	redisHost := getEnv("REDIS_HOST", "redis")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := getEnv("REDIS_PORT", "6379")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	return addr
}
