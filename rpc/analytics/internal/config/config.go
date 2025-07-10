package config

import (
	"os"
	"strconv"
	"time"
)

// Config конфигурация analytics сервиса с OpenTelemetry
type Config struct {
	HTTP          HTTPConfig
	GRPC          GRPCConfig
	OpenTelemetry OTelConfig
	Redis         RedisConfig
	Log           LogConfig
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type GRPCConfig struct {
	Address string
}

type OTelConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Tracing
	TracingEnabled    bool
	JaegerEndpoint    string
	ZipkinEndpoint    string
	OTLPTraceEndpoint string
	OTLPTraceInsecure bool

	// Metrics
	MetricsEnabled     bool
	PrometheusEndpoint string
	OTLPMetricEndpoint string
	OTLPMetricInsecure bool
	MetricsPort        string

	// Logging
	LoggingEnabled  bool
	OTLPLogEndpoint string
	OTLPLogInsecure bool

	// Sampling
	TraceSamplingRatio float64

	// Resource attributes
	ResourceAttributes map[string]string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type LogConfig struct {
	Level  string
	Format string // json, console
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
		GRPC: GRPCConfig{
			Address: getEnv("GRPC_ADDRESS", ":9090"),
		},
		OpenTelemetry: OTelConfig{
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "analytics-service"),
			ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("OTEL_ENVIRONMENT", "development"),

			// Tracing
			TracingEnabled:    getBool("OTEL_TRACING_ENABLED", true),
			JaegerEndpoint:    getEnv("OTEL_JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			ZipkinEndpoint:    getEnv("OTEL_ZIPKIN_ENDPOINT", "http://localhost:9411/api/v2/spans"),
			OTLPTraceEndpoint: getEnv("OTEL_OTLP_TRACE_ENDPOINT", "http://localhost:4317"),
			OTLPTraceInsecure: getBool("OTEL_OTLP_TRACE_INSECURE", true),

			// Metrics
			MetricsEnabled:     getBool("OTEL_METRICS_ENABLED", true),
			PrometheusEndpoint: getEnv("OTEL_PROMETHEUS_ENDPOINT", "http://localhost:9090"),
			OTLPMetricEndpoint: getEnv("OTEL_OTLP_METRIC_ENDPOINT", "http://localhost:4317"),
			OTLPMetricInsecure: getBool("OTEL_OTLP_METRIC_INSECURE", true),
			MetricsPort:        getEnv("OTEL_METRICS_PORT", "8081"),

			// Logging
			LoggingEnabled:  getBool("OTEL_LOGGING_ENABLED", true),
			OTLPLogEndpoint: getEnv("OTEL_OTLP_LOG_ENDPOINT", "http://localhost:4317"),
			OTLPLogInsecure: getBool("OTEL_OTLP_LOG_INSECURE", true),

			// Sampling
			TraceSamplingRatio: getFloat("OTEL_TRACE_SAMPLING_RATIO", 1.0),

			// Resource attributes
			ResourceAttributes: map[string]string{
				"service.name":           getEnv("OTEL_SERVICE_NAME", "analytics-service"),
				"service.version":        getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
				"deployment.environment": getEnv("OTEL_ENVIRONMENT", "development"),
			},
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
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

// getBool получает boolean из переменной окружения
func getBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getInt получает int из переменной окружения
func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// getFloat получает float64 из переменной окружения
func getFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
