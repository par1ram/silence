package app

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/adapters/database"
	grpcadapter "github.com/par1ram/silence/rpc/analytics/internal/adapters/grpc"
	"github.com/par1ram/silence/rpc/analytics/internal/config"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"github.com/par1ram/silence/rpc/analytics/internal/services"
	"go.uber.org/zap"
)

// App представляет приложение аналитики
// Основная структура, управляющая жизненным циклом сервиса
// Используйте New для создания, Start для запуска, Shutdown для остановки
type App struct {
	config          *config.Config
	logger          *zap.Logger
	grpcServer      *grpcadapter.Server
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

	metricsRepo, err := database.NewInfluxDBRepository(
		cfg.InfluxDB.URL,
		cfg.InfluxDB.Token,
		cfg.InfluxDB.Org,
		cfg.InfluxDB.Bucket,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics repository: %w", err)
	}

	analyticsSvc := services.NewAnalyticsService(
		metricsRepo,
		nil, // TODO: dashboard repository
		nil, // TODO: metrics collector
		nil, // TODO: alert service
		logger,
	)

	grpcServer := grpcadapter.NewServer(analyticsSvc, logger, cfg)

	return &App{
		config:          cfg,
		logger:          logger,
		grpcServer:      grpcServer,
		metricsRepo:     metricsRepo,
		analyticsSvc:    analyticsSvc,
		shutdownTimeout: 30 * time.Second,
	}, nil
}

// Start запускает gRPC сервер аналитики
func (a *App) Start() error {
	a.logger.Info("Starting analytics service",
		zap.String("grpc_address", a.config.GRPC.Address),
		zap.String("influxdb_url", a.config.InfluxDB.URL),
	)

	ctx := context.Background()
	return a.grpcServer.Start(ctx)
}

// Shutdown останавливает gRPC сервер и закрывает ресурсы
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down analytics service...")
	if err := a.grpcServer.Stop(); err != nil {
		a.logger.Error("Error shutting down gRPC server", zap.String("error", err.Error()))
	}
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
