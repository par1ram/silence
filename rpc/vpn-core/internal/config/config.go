package config

import (
	"os"
	"strconv"
)

// Config конфигурация VPN Core сервиса
type Config struct {
	HTTPPort     string
	GRPCPort     string
	LogLevel     string
	Version      string
	WireGuardDir string
	Interface    string
	ListenPort   int
	MTU          int
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", ":8082"),
		GRPCPort:     getEnv("GRPC_PORT", ":9092"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Version:      getEnv("VERSION", "1.0.0"),
		WireGuardDir: getEnv("WIREGUARD_DIR", "/etc/wireguard"),
		Interface:    getEnv("WIREGUARD_INTERFACE", "wg0"),
		ListenPort:   getEnvInt("WIREGUARD_LISTEN_PORT", 51820),
		MTU:          getEnvInt("WIREGUARD_MTU", 1420),
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
