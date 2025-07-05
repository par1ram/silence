package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/par1ram/silence/rpc/vpn-core/internal/adapters/grpc"
	"github.com/par1ram/silence/rpc/vpn-core/internal/adapters/http"
	"github.com/par1ram/silence/rpc/vpn-core/internal/adapters/wireguard"
	"github.com/par1ram/silence/rpc/vpn-core/internal/config"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"github.com/par1ram/silence/rpc/vpn-core/internal/services"
	"github.com/par1ram/silence/shared/logger"
	"go.uber.org/zap"
)

// App приложение VPN Core
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

// Run запускает приложение
func Run() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем логгер
	logger := logger.NewLogger("vpn-core")
	defer logger.Sync()

	// Создаем приложение
	app := NewApp(cfg, logger)

	// Создаем WireGuard адаптер (используем mock для тестирования)
	wgAdapter := wireguard.NewMockWGAdapter(logger)

	// Создаем сервисы
	healthService := services.NewHealthService("vpn-core", cfg.Version)
	keyGenerator := services.NewKeyGenerator()
	tunnelManager := services.NewTunnelService(keyGenerator, wgAdapter, logger)
	peerManager := services.NewPeerService(logger)

	// Создаем сервис мониторинга
	monitorService := services.NewMonitorService(tunnelManager, peerManager, wgAdapter, logger)

	// Создаем HTTP обработчики
	handlers := http.NewHandlers(healthService, tunnelManager, peerManager, logger)

	// Создаем HTTP сервер
	httpServer := http.NewServer(cfg.HTTPPort, handlers, logger)
	app.AddService(httpServer)

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer(cfg.GRPCPort, tunnelManager, peerManager, logger)
	app.AddService(grpcServer)

	// Добавляем сервис мониторинга
	app.AddService(&MonitorServiceWrapper{
		monitorService: monitorService,
		logger:         logger,
	})

	// Запускаем приложение
	app.run()
}

// MonitorServiceWrapper обертка для MonitorService для интеграции с App
type MonitorServiceWrapper struct {
	monitorService ports.MonitorService
	logger         *zap.Logger
}

func (m *MonitorServiceWrapper) Start(ctx context.Context) error {
	m.logger.Info("starting monitor service")
	return m.monitorService.StartMonitoring(ctx)
}

func (m *MonitorServiceWrapper) Stop(ctx context.Context) error {
	m.logger.Info("stopping monitor service")
	return m.monitorService.StopMonitoring(ctx)
}

func (m *MonitorServiceWrapper) Name() string {
	return "monitor"
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	for _, service := range a.services {
		a.logger.Info("stopping service", zap.String("service", service.Name()))
		if err := service.Stop(shutdownCtx); err != nil {
			a.logger.Error("failed to stop service", zap.String("service", service.Name()), zap.Error(err))
		}
	}

	a.logger.Info("application shutdown complete")
}

const shutdownTimeout = 30 // секунды
