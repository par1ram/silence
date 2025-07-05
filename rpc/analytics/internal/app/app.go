package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/analytics/internal/adapters/database"
	httpadapter "github.com/par1ram/silence/rpc/analytics/internal/adapters/http"
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
	httpServer      *http.Server
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

	router := mux.NewRouter()
	analyticsHandler := httpadapter.NewAnalyticsHandler(analyticsSvc, logger)
	analyticsHandler.RegisterRoutes(router)

	router.Use(LoggingMiddleware(logger))
	router.Use(CORSMiddleware())

	httpServer := &http.Server{
		Addr:         cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		config:          cfg,
		logger:          logger,
		httpServer:      httpServer,
		metricsRepo:     metricsRepo,
		analyticsSvc:    analyticsSvc,
		shutdownTimeout: 30 * time.Second,
	}, nil
}

// Start запускает HTTP сервер аналитики
func (a *App) Start() error {
	a.logger.Info("Starting analytics service",
		zap.String("http_port", a.config.HTTP.Port),
		zap.String("influxdb_url", a.config.InfluxDB.URL),
	)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error", zap.String("error", err.Error()))
		}
	}()
	a.logger.Info("Analytics service started successfully")
	return nil
}

// Shutdown останавливает HTTP сервер и закрывает ресурсы
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down analytics service...")
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("Error shutting down HTTP server", zap.String("error", err.Error()))
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
