package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/config"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/svc"
	"github.com/par1ram/silence/shared/logger"
	"go.uber.org/zap"
)

// App приложение DPI Bypass
type App struct {
	config   *config.Config
	logger   *zap.Logger
	services []Service
}

// Service интерфейс для сервисов приложения
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// NewApp создает новое приложение
func NewApp(cfg *config.Config, logger *zap.Logger) *App {
	return &App{
		config:   cfg,
		logger:   logger,
		services: make([]Service, 0),
	}
}

// AddService добавляет сервис в приложение
func (a *App) AddService(service Service) {
	a.services = append(a.services, service)
}

const defaultShutdownTimeout = 30 // секунд

// Run запускает приложение DPI Bypass
func Run() {
	cfg := config.Load()
	logger := logger.NewLogger("dpi-bypass")
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	app := NewApp(cfg, logger)

	// Создаем контекст сервиса
	svcCtx := svc.NewServiceContext(cfg, logger)

	// Добавляем gRPC сервер
	app.AddService(svcCtx.GRPCServer)

	// Запускаем приложение
	app.run()
}

// run запускает все сервисы с graceful shutdown
func (a *App) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем все сервисы
	for _, service := range a.services {
		go func(s Service) {
			a.logger.Info("starting service", zap.String("service", s.Name()))
			if err := s.Start(ctx); err != nil {
				a.logger.Error("service failed", zap.String("service", s.Name()), zap.Error(err))
			}
		}(service)
	}

	// Ждем сигнала завершения
	<-sigChan
	a.logger.Info("shutdown signal received")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer shutdownCancel()

	for _, service := range a.services {
		a.logger.Info("stopping service", zap.String("service", service.Name()))
		if err := service.Stop(shutdownCtx); err != nil {
			a.logger.Error("failed to stop service", zap.String("service", service.Name()), zap.Error(err))
		}
	}

	a.logger.Info("application shutdown complete")
}
