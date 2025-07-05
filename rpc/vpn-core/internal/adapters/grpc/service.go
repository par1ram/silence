package grpc

import (
	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
)

// VpnCoreService реализация gRPC сервиса
type VpnCoreService struct {
	proto.UnimplementedVpnCoreServiceServer
	tunnelManager ports.TunnelManager
	peerManager   ports.PeerManager
	logger        *zap.Logger
}

// NewVpnCoreService конструктор
func NewVpnCoreService(tunnelManager ports.TunnelManager, peerManager ports.PeerManager, logger *zap.Logger) *VpnCoreService {
	return &VpnCoreService{
		tunnelManager: tunnelManager,
		peerManager:   peerManager,
		logger:        logger,
	}
}
