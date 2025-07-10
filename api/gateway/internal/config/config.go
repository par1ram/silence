package config

import (
	"os"
	"strconv"
)

// Config конфигурация приложения
type Config struct {
	HTTPPort         string
	LogLevel         string
	Version          string
	AuthURL          string
	VPNCoreURL       string
	DPIBypassURL     string
	AnalyticsURL     string
	ServerManagerURL string
	NotificationsURL string
	JWTSecret        string
	InternalAPIToken string

	// gRPC service URLs
	AuthGRPCURL          string
	VPNCoreGRPCURL       string
	DPIBypassGRPCURL     string
	AnalyticsGRPCURL     string
	ServerManagerGRPCURL string
	NotificationsGRPCURL string

	// Rate Limiting настройки
	RateLimitEnabled bool
	RateLimitRPS     int // requests per second
	RateLimitBurst   int // burst size
	RateLimitWindow  int // window size in seconds
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Version:          getEnv("VERSION", "1.0.0"),
		AuthURL:          getEnv("AUTH_URL", "http://localhost:8081"),
		VPNCoreURL:       getEnv("VPN_CORE_URL", "http://localhost:8082"),
		DPIBypassURL:     getEnv("DPI_BYPASS_URL", "http://localhost:8083"),
		AnalyticsURL:     getEnv("ANALYTICS_URL", "http://localhost:8084"),
		ServerManagerURL: getEnv("SERVER_MANAGER_URL", "http://localhost:8085"),
		NotificationsURL: getEnv("NOTIFICATIONS_URL", "http://localhost:8087"),
		JWTSecret:        getEnv("JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		InternalAPIToken: getEnv("INTERNAL_API_TOKEN", "super-secret-internal-token"),

		// gRPC service URLs
		AuthGRPCURL:          getEnv("AUTH_GRPC_SERVICE_URL", "localhost:9081"),
		VPNCoreGRPCURL:       getEnv("VPN_CORE_GRPC_SERVICE_URL", "localhost:9082"),
		DPIBypassGRPCURL:     getEnv("DPI_BYPASS_GRPC_SERVICE_URL", "localhost:9083"),
		AnalyticsGRPCURL:     getEnv("ANALYTICS_GRPC_SERVICE_URL", "localhost:9084"),
		ServerManagerGRPCURL: getEnv("SERVER_MANAGER_GRPC_SERVICE_URL", "localhost:9085"),
		NotificationsGRPCURL: getEnv("NOTIFICATIONS_GRPC_SERVICE_URL", "localhost:9087"),

		// Rate Limiting настройки
		RateLimitEnabled: getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRPS:     getEnvInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst:   getEnvInt("RATE_LIMIT_BURST", 200),
		RateLimitWindow:  getEnvInt("RATE_LIMIT_WINDOW", 60),
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

// getEnvBool получает булево значение переменной окружения
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}
