package config

import (
	"os"
	"strconv"
	"time"
)

// Config конфигурация сервиса уведомлений
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	RabbitMQ  RabbitMQConfig  `json:"rabbitmq"`
	Email     EmailConfig     `json:"email"`
	SMS       SMSConfig       `json:"sms"`
	Push      PushConfig      `json:"push"`
	Telegram  TelegramConfig  `json:"telegram"`
	Slack     SlackConfig     `json:"slack"`
	Redis     RedisConfig     `json:"redis"`
	Logging   LoggingConfig   `json:"logging"`
	Analytics AnalyticsConfig `json:"analytics"`
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
}

// RabbitMQConfig конфигурация RabbitMQ
type RabbitMQConfig struct {
	URL           string `json:"url"`
	Exchange      string `json:"exchange"`
	Queue         string `json:"queue"`
	RoutingKey    string `json:"routing_key"`
	ConsumerTag   string `json:"consumer_tag"`
	PrefetchCount int    `json:"prefetch_count"`
}

// EmailConfig конфигурация email
type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

// SMSConfig конфигурация SMS
type SMSConfig struct {
	Provider   string `json:"provider"` // twilio, nexmo, etc.
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	From       string `json:"from"`
}

// PushConfig конфигурация push уведомлений
type PushConfig struct {
	Provider string `json:"provider"` // fcm, apns, etc.
	APIKey   string `json:"api_key"`
	AppID    string `json:"app_id"`
}

// TelegramConfig конфигурация Telegram
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// SlackConfig конфигурация Slack
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
}

// RedisConfig конфигурация Redis
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LoggingConfig конфигурация логирования
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// AnalyticsConfig конфигурация интеграции с аналитикой
type AnalyticsConfig struct {
	URL string `json:"url"`
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("NOTIFICATIONS_PORT", "8080"),
			ReadTimeout:  getDurationEnv("NOTIFICATIONS_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("NOTIFICATIONS_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("NOTIFICATIONS_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("NOTIFICATIONS_DB_HOST", "localhost"),
			Port:     getEnv("NOTIFICATIONS_DB_PORT", "5432"),
			User:     getEnv("NOTIFICATIONS_DB_USER", "notifications"),
			Password: getEnv("NOTIFICATIONS_DB_PASSWORD", "notifications"),
			DBName:   getEnv("NOTIFICATIONS_DB_NAME", "notifications"),
			SSLMode:  getEnv("NOTIFICATIONS_DB_SSLMODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:           getEnv("NOTIFICATIONS_RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:      getEnv("NOTIFICATIONS_RABBITMQ_EXCHANGE", "notifications"),
			Queue:         getEnv("NOTIFICATIONS_RABBITMQ_QUEUE", "notifications"),
			RoutingKey:    getEnv("NOTIFICATIONS_RABBITMQ_ROUTING_KEY", "notification"),
			ConsumerTag:   getEnv("NOTIFICATIONS_RABBITMQ_CONSUMER_TAG", "notifications-consumer"),
			PrefetchCount: getIntEnv("NOTIFICATIONS_RABBITMQ_PREFETCH_COUNT", 10),
		},
		Email: EmailConfig{
			Host:     getEnv("NOTIFICATIONS_EMAIL_HOST", "localhost"),
			Port:     getIntEnv("NOTIFICATIONS_EMAIL_PORT", 587),
			Username: getEnv("NOTIFICATIONS_EMAIL_USERNAME", ""),
			Password: getEnv("NOTIFICATIONS_EMAIL_PASSWORD", ""),
			From:     getEnv("NOTIFICATIONS_EMAIL_FROM", "noreply@silence.com"),
			UseTLS:   getBoolEnv("NOTIFICATIONS_EMAIL_USE_TLS", true),
		},
		SMS: SMSConfig{
			Provider:   getEnv("NOTIFICATIONS_SMS_PROVIDER", "twilio"),
			AccountSID: getEnv("NOTIFICATIONS_SMS_ACCOUNT_SID", ""),
			AuthToken:  getEnv("NOTIFICATIONS_SMS_AUTH_TOKEN", ""),
			From:       getEnv("NOTIFICATIONS_SMS_FROM", ""),
		},
		Push: PushConfig{
			Provider: getEnv("NOTIFICATIONS_PUSH_PROVIDER", "fcm"),
			APIKey:   getEnv("NOTIFICATIONS_PUSH_API_KEY", ""),
			AppID:    getEnv("NOTIFICATIONS_PUSH_APP_ID", ""),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("NOTIFICATIONS_TELEGRAM_BOT_TOKEN", ""),
			ChatID:   getEnv("NOTIFICATIONS_TELEGRAM_CHAT_ID", ""),
		},
		Slack: SlackConfig{
			WebhookURL: getEnv("NOTIFICATIONS_SLACK_WEBHOOK_URL", ""),
			Channel:    getEnv("NOTIFICATIONS_SLACK_CHANNEL", "#general"),
			Username:   getEnv("NOTIFICATIONS_SLACK_USERNAME", "Silence Notifications"),
		},
		Redis: RedisConfig{
			Host:     getEnv("NOTIFICATIONS_REDIS_HOST", "localhost"),
			Port:     getEnv("NOTIFICATIONS_REDIS_PORT", "6379"),
			Password: getEnv("NOTIFICATIONS_REDIS_PASSWORD", ""),
			DB:       getIntEnv("NOTIFICATIONS_REDIS_DB", 0),
		},
		Logging: LoggingConfig{
			Level:  getEnv("NOTIFICATIONS_LOG_LEVEL", "info"),
			Format: getEnv("NOTIFICATIONS_LOG_FORMAT", "json"),
		},
		Analytics: AnalyticsConfig{
			URL: getEnv("NOTIFICATIONS_ANALYTICS_URL", "http://localhost:8084"),
		},
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv получает целочисленное значение переменной окружения
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getBoolEnv получает булево значение переменной окружения
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getDurationEnv получает значение времени из переменной окружения
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
