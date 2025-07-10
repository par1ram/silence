package app

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/adapters"
	"github.com/par1ram/silence/rpc/analytics/internal/adapters/database"
	"github.com/par1ram/silence/rpc/analytics/internal/config"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"github.com/par1ram/silence/rpc/analytics/internal/services"
	"github.com/par1ram/silence/rpc/analytics/internal/telemetry"
	"github.com/par1ram/silence/shared/redis"
	"go.uber.org/zap"
)

// App представляет приложение аналитики
// Основная структура, управляющая жизненным циклом сервиса
// Используйте New для создания, Start для запуска, Shutdown для остановки
type App struct {
	config          *config.Config
	logger          *zap.Logger
	redisClient     *redis.Client
	clickhouseRepo  ports.MetricsRepository
	influxdbRepo    ports.MetricsRepository
	metricsRepo     ports.MetricsRepository
	analyticsSvc    ports.AnalyticsService
	shutdownTimeout time.Duration
}

// New создает новое приложение аналитики
func New(logger *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Создаем Redis клиент
	redisClient, err := redis.NewClient(&redis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Создаем ClickHouse репозиторий
	clickhouseRepo, err := database.NewClickHouseRepository(
		cfg.ClickHouse.Host,
		cfg.ClickHouse.Port,
		cfg.ClickHouse.Database,
		cfg.ClickHouse.Username,
		cfg.ClickHouse.Password,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse repository: %w", err)
	}

	// Создаем InfluxDB репозиторий
	influxdbRepo, err := database.NewInfluxDBRepository(
		cfg.InfluxDB.URL,
		cfg.InfluxDB.Token,
		cfg.InfluxDB.Org,
		cfg.InfluxDB.Bucket,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create influxdb repository: %w", err)
	}

	// Создаем репозитории и сервисы
	metricsRepo := adapters.NewRedisMetricsRepository(redisClient.GetClient(), logger)
	dashboardRepo := adapters.NewRedisDashboardRepository(redisClient.GetClient(), logger)
	metricsCollector := adapters.NewRedisMetricsCollector(redisClient.GetClient(), logger)
	alertService := adapters.NewRedisAlertService(redisClient.GetClient(), logger)

	// Создаем telemetry менеджер
	telemetryManager, err := telemetry.NewTelemetryManager(cfg.OpenTelemetry, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
	}

	// Создаем tracing менеджер
	tracingManager := telemetry.NewTracingManager(
		telemetryManager.GetTracer(),
		logger,
	)

	// Создаем telemetry metrics collector
	telemetryMetricsCollector, err := telemetry.NewMetricsCollector(
		telemetryManager.GetMeter(),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry metrics collector: %w", err)
	}

	analyticsSvc := services.NewAnalyticsService(
		clickhouseRepo, // Используем ClickHouse как основной репозиторий
		dashboardRepo,
		metricsCollector,
		alertService,
		logger,
		telemetryMetricsCollector,
		tracingManager,
	)

	return &App{
		config:          cfg,
		logger:          logger,
		redisClient:     redisClient,
		clickhouseRepo:  clickhouseRepo,
		influxdbRepo:    influxdbRepo,
		metricsRepo:     metricsRepo,
		analyticsSvc:    analyticsSvc,
		shutdownTimeout: 30 * time.Second,
	}, nil
}

// Start запускает gRPC сервер аналитики
func (a *App) Start() error {
	a.logger.Info("Starting analytics service",
		zap.String("grpc_address", a.config.GRPC.Address),
		zap.String("redis_address", a.config.Redis.Address),
	)

	// В реальной реализации здесь должен быть запуск gRPC сервера
	// Пока что просто возвращаем nil
	return nil
}

// Shutdown останавливает gRPC сервер и закрывает ресурсы
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down analytics service...")

	// Закрываем Redis клиент
	if err := a.redisClient.Close(); err != nil {
		a.logger.Error("Error closing redis client", zap.Error(err))
	}

	// Закрываем ClickHouse репозиторий
	if closer, ok := a.clickhouseRepo.(interface{ Close() }); ok {
		closer.Close()
	}

	// Закрываем InfluxDB репозиторий
	if closer, ok := a.influxdbRepo.(interface{ Close() }); ok {
		closer.Close()
	}

	// Закрываем другие ресурсы
	if closer, ok := a.metricsRepo.(interface{ Close() }); ok {
		closer.Close()
	}

	a.logger.Info("Analytics service stopped")
	return nil
}

// ShutdownTimeout возвращает таймаут для graceful shutdown
func (a *App) ShutdownTimeout() time.Duration {
	return a.shutdownTimeout
}
