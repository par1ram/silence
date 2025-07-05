package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config конфигурация сервиса управления серверами
type Config struct {
	// HTTP сервер
	HTTPPort         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration

	// gRPC сервер
	GRPCPort string

	// Версия сервиса
	Version string

	// База данных
	Database DatabaseConfig

	// Docker
	Docker DockerConfig

	// Оркестратор
	Orchestrator OrchestratorConfig

	// Мониторинг
	Monitoring MonitoringConfig

	// Географическое распределение
	Geographic GeographicConfig
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DockerConfig конфигурация Docker
type DockerConfig struct {
	Host       string
	APIVersion string
	Timeout    time.Duration
}

// OrchestratorConfig конфигурация оркестратора
type OrchestratorConfig struct {
	Type       string // "docker" или "kubernetes"
	Kubeconfig string // путь к kubeconfig файлу
	Namespace  string // namespace для Kubernetes
}

// MonitoringConfig конфигурация мониторинга
type MonitoringConfig struct {
	HealthCheckInterval time.Duration
	MetricsInterval     time.Duration
	AlertThreshold      float64
}

// GeographicConfig конфигурация географического распределения
type GeographicConfig struct {
	DefaultRegion string
	Regions       []string
	AutoScaling   bool
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	// Загружаем .env файл если существует
	_ = godotenv.Load()

	config := &Config{
		HTTPPort:         getEnv("HTTP_PORT", "8085"),
		HTTPReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 30*time.Second),
		HTTPWriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		HTTPIdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		GRPCPort:         getEnv("GRPC_PORT", "50055"),
		Version:          getEnv("VERSION", "1.0.0"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "silence_server_manager"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		Docker: DockerConfig{
			Host:       getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
			APIVersion: getEnv("DOCKER_API_VERSION", "1.41"),
			Timeout:    getEnvDuration("DOCKER_TIMEOUT", 30*time.Second),
		},

		Orchestrator: OrchestratorConfig{
			Type:       getEnv("ORCHESTRATOR_TYPE", "docker"),
			Kubeconfig: getEnv("KUBECONFIG", ""),
			Namespace:  getEnv("KUBERNETES_NAMESPACE", "default"),
		},

		Monitoring: MonitoringConfig{
			HealthCheckInterval: getEnvDuration("HEALTH_CHECK_INTERVAL", 30*time.Second),
			MetricsInterval:     getEnvDuration("METRICS_INTERVAL", 60*time.Second),
			AlertThreshold:      getEnvFloat("ALERT_THRESHOLD", 0.8),
		},

		Geographic: GeographicConfig{
			DefaultRegion: getEnv("DEFAULT_REGION", "us-east-1"),
			Regions:       getEnvSlice("REGIONS", []string{"us-east-1", "us-west-1", "eu-west-1"}),
			AutoScaling:   getEnvBool("AUTO_SCALING", true),
		},
	}

	return config
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает целочисленную переменную окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloat получает переменную окружения типа float
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvBool получает булеву переменную окружения
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration получает переменную окружения типа duration
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvSlice получает слайс строк из переменной окружения
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Простая реализация - разделение по запятой
		// В продакшене можно использовать более сложную логику
		return []string{value} // Упрощенно для примера
	}
	return defaultValue
}
