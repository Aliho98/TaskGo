package config

import (
	"os"
	"time"
)

type Config struct {
	Env             string
	HTTPPort        string
	PostgresHost    string
	RedisAddr       string
	RedisPassword   string
	ShutdownTimeout time.Duration
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
func LoadConfig() Config {
	return Config{
		Env:           getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		PostgresHost:  getEnv("POSTGRES_HOST", "postgres://postgres:postgres@localhost:5432/taskdb?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}
}
