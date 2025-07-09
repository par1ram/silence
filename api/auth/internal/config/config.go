package config

import (
	"os"
	"strconv"
	"time"
)

// Config конфигурация auth сервиса
type Config struct {
	HTTPPort         string
	GRPCPort         int
	LogLevel         string
	Version          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	JWTExpiresIn     time.Duration
	InternalAPIToken string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		HTTPPort:         getEnv("HTTP_PORT", "9999"),
		GRPCPort:         getEnvInt("GRPC_PORT", 9998),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Version:          getEnv("VERSION", "1.0.0"),
		DBHost:           getEnv("DB_HOST", "127.0.0.1"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "pariram"),
		DBPassword:       getEnv("DB_PASSWORD", "password"),
		DBName:           getEnv("DB_NAME", "silence_auth"),
		JWTSecret:        getEnv("JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		JWTExpiresIn:     time.Duration(getEnvInt("JWT_EXPIRES_IN_HOURS", 24)) * time.Hour,
		InternalAPIToken: getEnv("INTERNAL_API_TOKEN", "super-secret-internal-token"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает целочисленное значение переменной окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
