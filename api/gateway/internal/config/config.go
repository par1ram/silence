package config

import (
	"os"
	"strconv"
)

// Config конфигурация приложения
type Config struct {
	HTTPPort     string
	LogLevel     string
	Version      string
	AuthURL      string
	VPNCoreURL   string
	DPIBypassURL string
	JWTSecret    string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", ":8080"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Version:      getEnv("VERSION", "1.0.0"),
		AuthURL:      getEnv("AUTH_URL", "http://localhost:8081"),
		VPNCoreURL:   getEnv("VPN_CORE_URL", "http://localhost:8082"),
		DPIBypassURL: getEnv("DPI_BYPASS_URL", "http://localhost:8083"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
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
