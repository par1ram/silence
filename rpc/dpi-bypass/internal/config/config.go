package config

import (
	"os"
)

type Config struct {
	HTTPPort string
	LogLevel string
	Version  string
}

func Load() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Version:  getEnv("VERSION", "1.0.0"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
