package config

import (
	"os"
	"time"
)

// Config конфигурация analytics сервиса
type Config struct {
	HTTP     HTTPConfig
	InfluxDB InfluxDBConfig
	Log      LogConfig
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type InfluxDBConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

type LogConfig struct {
	Level string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	return &Config{
		HTTP: HTTPConfig{
			Port:         getEnv("HTTP_PORT", "8080"),
			ReadTimeout:  getDuration("HTTP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		InfluxDB: InfluxDBConfig{
			URL:    getEnv("INFLUXDB_URL", "http://localhost:8086"),
			Token:  getEnv("INFLUXDB_TOKEN", ""),
			Org:    getEnv("INFLUXDB_ORG", "silence"),
			Bucket: getEnv("INFLUXDB_BUCKET", "analytics"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDuration получает duration из переменной окружения
func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
