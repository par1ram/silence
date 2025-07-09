package svc

import (
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/bypass"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/grpc"
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
	BypassService ports.DPIBypassService
	BypassAdapter ports.BypassAdapter
	GRPCServer    *grpc.Server
}

// NewServiceContext создает новый контекст сервиса
func NewServiceContext(cfg *config.Config, logger *zap.Logger) *ServiceContext {
	// Создаем сервисы
	healthService := services.NewHealthService("dpi-bypass", cfg.Version)

	// Создаем мульти-адаптер для обфускации
	bypassAdapter := bypass.NewMultiBypassAdapter(logger)

	// Создаем bypass сервис
	bypassService := services.NewBypassService(bypassAdapter, logger)

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer(bypassService, logger, cfg)

	return &ServiceContext{
		Config:        cfg,
		Logger:        logger,
		HealthService: healthService,
		BypassService: bypassService,
		BypassAdapter: bypassAdapter,
		GRPCServer:    grpcServer,
	}
}
