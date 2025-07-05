package grpc

import (
	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// domainTunnelToProto конвертирует доменную модель туннеля в proto
func (s *VpnCoreService) domainTunnelToProto(tunnel *domain.Tunnel) *proto.Tunnel {
	protoTunnel := &proto.Tunnel{
		Id:               tunnel.ID,
		Name:             tunnel.Name,
		Interface:        tunnel.Interface,
		Status:           s.domainTunnelStatusToProto(tunnel.Status),
		PublicKey:        tunnel.PublicKey,
		PrivateKey:       tunnel.PrivateKey,
		ListenPort:       int32(tunnel.ListenPort),
		Mtu:              int32(tunnel.MTU),
		CreatedAt:        timestamppb.New(tunnel.CreatedAt),
		UpdatedAt:        timestamppb.New(tunnel.UpdatedAt),
		AutoRecovery:     tunnel.AutoRecovery,
		RecoveryAttempts: int32(tunnel.RecoveryAttempts),
	}

	// Добавляем новые поля для мониторинга
	if !tunnel.LastHealthCheck.IsZero() {
		protoTunnel.LastHealthCheck = timestamppb.New(tunnel.LastHealthCheck)
	}
	if tunnel.HealthStatus != "" {
		protoTunnel.HealthStatus = tunnel.HealthStatus
	}

	return protoTunnel
}

// domainTunnelStatusToProto конвертирует статус туннеля
func (s *VpnCoreService) domainTunnelStatusToProto(status domain.TunnelStatus) proto.TunnelStatus {
	switch status {
	case domain.TunnelStatusInactive:
		return proto.TunnelStatus_TUNNEL_STATUS_INACTIVE
	case domain.TunnelStatusActive:
		return proto.TunnelStatus_TUNNEL_STATUS_ACTIVE
	case domain.TunnelStatusError:
		return proto.TunnelStatus_TUNNEL_STATUS_ERROR
	case domain.TunnelStatusRecovering:
		return proto.TunnelStatus_TUNNEL_STATUS_RECOVERING
	default:
		return proto.TunnelStatus_TUNNEL_STATUS_UNSPECIFIED
	}
}
