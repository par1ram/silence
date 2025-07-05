package svc

import (
	"github.com/par1ram/silence/api/gateway/internal/adapters/http"
	"github.com/par1ram/silence/api/gateway/internal/config"
	"github.com/par1ram/silence/api/gateway/internal/ports"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

// ServiceContext контекст сервиса с зависимостями
type ServiceContext struct {
	Config        *config.Config
	Logger        *zap.Logger
	HealthService ports.HealthService
	ProxyService  ports.ProxyService
	HTTPServer    *http.Server
}

// NewServiceContext создает новый контекст сервиса
func NewServiceContext(cfg *config.Config, logger *zap.Logger) *ServiceContext {
	// Создаем сервисы
	healthService := services.NewHealthService("gateway", cfg.Version)
	proxyService := services.NewProxyService(cfg.AuthURL, cfg.VPNCoreURL, logger)

	// Создаем HTTP обработчики
	handlers := http.NewHandlers(healthService, proxyService, logger)

	// Создаем HTTP сервер
	httpServer := http.NewServer(cfg.HTTPPort, handlers, logger, cfg.JWTSecret)

	return &ServiceContext{
		Config:        cfg,
		Logger:        logger,
		HealthService: healthService,
		ProxyService:  proxyService,
		HTTPServer:    httpServer,
	}
}
