package http

import (
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
)

// Handlers HTTP обработчики
// (реализация обработчиков вынесена в отдельные файлы)
type Handlers struct {
	healthService ports.HealthService
	tunnelManager ports.TunnelManager
	peerManager   ports.PeerManager
	logger        *zap.Logger
}

// NewHandlers создает новые HTTP обработчики
func NewHandlers(healthService ports.HealthService, tunnelManager ports.TunnelManager, peerManager ports.PeerManager, logger *zap.Logger) *Handlers {
	return &Handlers{
		healthService: healthService,
		tunnelManager: tunnelManager,
		peerManager:   peerManager,
		logger:        logger,
	}
}
