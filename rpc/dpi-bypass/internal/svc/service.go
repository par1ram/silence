package svc

import (
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/bypass"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/http"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/config"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/ports"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/services"
	"go.uber.org/zap"
)

// ServiceContext контекст сервиса с зависимостями
type ServiceContext struct {
	Config        *config.Config
	Logger        *zap.Logger
	HealthService ports.HealthService
	BypassService ports.BypassService
	BypassAdapter ports.BypassAdapter
	HTTPServer    *http.Server
}

// NewServiceContext создает новый контекст сервиса
func NewServiceContext(cfg *config.Config, logger *zap.Logger) *ServiceContext {
	// Создаем сервисы
	healthService := services.NewHealthService("dpi-bypass", cfg.Version)

	// Создаем мульти-адаптер для обфускации
	bypassAdapter := bypass.NewMultiBypassAdapter(logger)

	// Создаем bypass сервис
	bypassService := services.NewBypassService(bypassAdapter, logger)

	// Создаем HTTP обработчики
	handlers := http.NewHandlers(healthService, bypassService, logger)

	// Создаем HTTP сервер
	httpServer := http.NewServer(cfg.HTTPPort, handlers, logger)

	return &ServiceContext{
		Config:        cfg,
		Logger:        logger,
		HealthService: healthService,
		BypassService: bypassService,
		BypassAdapter: bypassAdapter,
		HTTPServer:    httpServer,
	}
}
