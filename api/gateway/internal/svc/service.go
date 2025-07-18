package svc

import (
	"context"
	"time"

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
	GRPCClients   *services.GRPCClients
	HTTPServer    *http.Server
}

// NewServiceContext создает новый контекст сервиса
func NewServiceContext(cfg *config.Config, logger *zap.Logger) *ServiceContext {
	// Создаем сервисы
	healthService := services.NewHealthService("gateway", cfg.Version)
	proxyService := services.NewProxyService(cfg.AuthURL, cfg.VPNCoreURL, cfg.DPIBypassURL, cfg.AnalyticsURL, cfg.ServerManagerURL, cfg.NotificationsURL, cfg.InternalAPIToken, logger)

	// Создаем gRPC клиенты
	grpcClients := services.NewGRPCClients(cfg, logger)

	// Инициализируем gRPC клиенты
	ctx := context.Background()
	if err := grpcClients.Initialize(ctx); err != nil {
		logger.Error("failed to initialize gRPC clients", zap.Error(err))
		// Продолжаем работу с HTTP прокси в случае ошибки
	}

	// Создаем rate limiter если включен
	var rateLimiter *http.RateLimiter
	if cfg.RateLimitEnabled {
		rateLimiter = http.NewRateLimiter(
			float64(cfg.RateLimitRPS),
			cfg.RateLimitBurst,
			time.Duration(cfg.RateLimitWindow)*time.Second,
			logger,
		)
		logger.Info("rate limiting enabled",
			zap.Int("rps", cfg.RateLimitRPS),
			zap.Int("burst", cfg.RateLimitBurst),
			zap.Int("window", cfg.RateLimitWindow),
		)
	} else {
		logger.Info("rate limiting disabled")
	}

	// Создаем HTTP обработчики
	handlers := http.NewHandlers(healthService, proxyService, rateLimiter, logger)

	// Создаем HTTP сервер
	httpServer := http.NewServer(cfg.HTTPPort, handlers, logger, cfg.JWTSecret, rateLimiter)

	return &ServiceContext{
		Config:        cfg,
		Logger:        logger,
		HealthService: healthService,
		ProxyService:  proxyService,
		GRPCClients:   grpcClients,
		HTTPServer:    httpServer,
	}
}
