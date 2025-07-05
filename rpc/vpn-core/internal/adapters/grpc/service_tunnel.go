package grpc

import (
	"context"
	"fmt"

	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTunnel создает новый туннель
func (s *VpnCoreService) CreateTunnel(ctx context.Context, req *proto.CreateTunnelRequest) (*proto.Tunnel, error) {
	s.logger.Info("creating tunnel", zap.String("name", req.Name))

	domainReq := &domain.CreateTunnelRequest{
		Name:         req.Name,
		ListenPort:   int(req.ListenPort),
		MTU:          int(req.Mtu),
		AutoRecovery: req.AutoRecovery,
	}

	tunnel, err := s.tunnelManager.CreateTunnel(ctx, domainReq)
	if err != nil {
		s.logger.Error("failed to create tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to create tunnel: %w", err)
	}

	return s.domainTunnelToProto(tunnel), nil
}

// GetTunnel получает туннель по ID
func (s *VpnCoreService) GetTunnel(ctx context.Context, req *proto.GetTunnelRequest) (*proto.Tunnel, error) {
	s.logger.Debug("getting tunnel", zap.String("id", req.Id))

	tunnel, err := s.tunnelManager.GetTunnel(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to get tunnel: %w", err)
	}

	return s.domainTunnelToProto(tunnel), nil
}

// ListTunnels возвращает список всех туннелей
func (s *VpnCoreService) ListTunnels(ctx context.Context, req *proto.ListTunnelsRequest) (*proto.ListTunnelsResponse, error) {
	s.logger.Debug("listing tunnels")

	tunnels, err := s.tunnelManager.ListTunnels(ctx)
	if err != nil {
		s.logger.Error("failed to list tunnels", zap.Error(err))
		return nil, fmt.Errorf("failed to list tunnels: %w", err)
	}

	protoTunnels := make([]*proto.Tunnel, len(tunnels))
	for i, tunnel := range tunnels {
		protoTunnels[i] = s.domainTunnelToProto(tunnel)
	}

	return &proto.ListTunnelsResponse{
		Tunnels: protoTunnels,
	}, nil
}

// DeleteTunnel удаляет туннель
func (s *VpnCoreService) DeleteTunnel(ctx context.Context, req *proto.DeleteTunnelRequest) (*proto.DeleteTunnelResponse, error) {
	s.logger.Info("deleting tunnel", zap.String("id", req.Id))

	err := s.tunnelManager.DeleteTunnel(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to delete tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to delete tunnel: %w", err)
	}

	return &proto.DeleteTunnelResponse{
		Success: true,
	}, nil
}

// StartTunnel запускает туннель
func (s *VpnCoreService) StartTunnel(ctx context.Context, req *proto.StartTunnelRequest) (*proto.StartTunnelResponse, error) {
	s.logger.Info("starting tunnel", zap.String("id", req.Id))

	err := s.tunnelManager.StartTunnel(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to start tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	return &proto.StartTunnelResponse{
		Success: true,
	}, nil
}

// StopTunnel останавливает туннель
func (s *VpnCoreService) StopTunnel(ctx context.Context, req *proto.StopTunnelRequest) (*proto.StopTunnelResponse, error) {
	s.logger.Info("stopping tunnel", zap.String("id", req.Id))

	err := s.tunnelManager.StopTunnel(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to stop tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to stop tunnel: %w", err)
	}

	return &proto.StopTunnelResponse{
		Success: true,
	}, nil
}

// GetTunnelStats получает статистику туннеля
func (s *VpnCoreService) GetTunnelStats(ctx context.Context, req *proto.GetTunnelStatsRequest) (*proto.TunnelStats, error) {
	s.logger.Debug("getting tunnel stats", zap.String("id", req.Id))

	stats, err := s.tunnelManager.GetTunnelStats(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get tunnel stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get tunnel stats: %w", err)
	}

	return &proto.TunnelStats{
		TunnelId:      stats.TunnelID,
		BytesRx:       stats.BytesRx,
		BytesTx:       stats.BytesTx,
		PeersCount:    int32(stats.PeersCount),
		ActivePeers:   int32(stats.ActivePeers),
		LastUpdated:   timestamppb.New(stats.LastUpdated),
		Uptime:        int64(stats.Uptime.Seconds()),
		ErrorCount:    int32(stats.ErrorCount),
		RecoveryCount: int32(stats.RecoveryCount),
	}, nil
}

// HealthCheck проверяет здоровье туннеля
func (s *VpnCoreService) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	s.logger.Debug("health check", zap.String("tunnel_id", req.TunnelId))

	domainReq := &domain.HealthCheckRequest{
		TunnelID: req.TunnelId,
	}

	health, err := s.tunnelManager.HealthCheck(ctx, domainReq)
	if err != nil {
		s.logger.Error("health check failed", zap.Error(err))
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	// Конвертируем здоровье пиров
	peersHealth := make([]*proto.PeerHealth, len(health.PeersHealth))
	for i, peerHealth := range health.PeersHealth {
		// Конвертируем статус
		var status proto.PeerStatus
		switch peerHealth.Status {
		case domain.PeerStatusActive:
			status = proto.PeerStatus_PEER_STATUS_ACTIVE
		case domain.PeerStatusInactive:
			status = proto.PeerStatus_PEER_STATUS_INACTIVE
		case domain.PeerStatusError:
			status = proto.PeerStatus_PEER_STATUS_ERROR
		case domain.PeerStatusOffline:
			status = proto.PeerStatus_PEER_STATUS_OFFLINE
		default:
			status = proto.PeerStatus_PEER_STATUS_UNSPECIFIED
		}

		peersHealth[i] = &proto.PeerHealth{
			PeerId:            peerHealth.PeerID,
			Status:            status,
			LastHandshake:     timestamppb.New(peerHealth.LastHandshake),
			Latency:           int64(peerHealth.Latency.Milliseconds()),
			PacketLoss:        peerHealth.PacketLoss,
			ConnectionQuality: peerHealth.ConnectionQuality,
		}
	}

	return &proto.HealthCheckResponse{
		TunnelId:    health.TunnelID,
		Status:      health.Status,
		LastCheck:   timestamppb.New(health.LastCheck),
		PeersHealth: peersHealth,
		Uptime:      int64(health.Uptime.Seconds()),
		ErrorCount:  int32(health.ErrorCount),
	}, nil
}

// EnableAutoRecovery включает автоматическое восстановление для туннеля
func (s *VpnCoreService) EnableAutoRecovery(ctx context.Context, req *proto.EnableAutoRecoveryRequest) (*proto.EnableAutoRecoveryResponse, error) {
	s.logger.Info("enabling auto recovery", zap.String("tunnel_id", req.TunnelId))

	err := s.tunnelManager.EnableAutoRecovery(ctx, req.TunnelId)
	if err != nil {
		s.logger.Error("failed to enable auto recovery", zap.Error(err))
		return nil, fmt.Errorf("failed to enable auto recovery: %w", err)
	}

	return &proto.EnableAutoRecoveryResponse{
		Success: true,
	}, nil
}

// DisableAutoRecovery отключает автоматическое восстановление для туннеля
func (s *VpnCoreService) DisableAutoRecovery(ctx context.Context, req *proto.DisableAutoRecoveryRequest) (*proto.DisableAutoRecoveryResponse, error) {
	s.logger.Info("disabling auto recovery", zap.String("tunnel_id", req.TunnelId))

	err := s.tunnelManager.DisableAutoRecovery(ctx, req.TunnelId)
	if err != nil {
		s.logger.Error("failed to disable auto recovery", zap.Error(err))
		return nil, fmt.Errorf("failed to disable auto recovery: %w", err)
	}

	return &proto.DisableAutoRecoveryResponse{
		Success: true,
	}, nil
}

// RecoverTunnel восстанавливает туннель
func (s *VpnCoreService) RecoverTunnel(ctx context.Context, req *proto.RecoverTunnelRequest) (*proto.RecoverTunnelResponse, error) {
	s.logger.Info("recovering tunnel", zap.String("tunnel_id", req.TunnelId))

	err := s.tunnelManager.RecoverTunnel(ctx, req.TunnelId)
	if err != nil {
		s.logger.Error("failed to recover tunnel", zap.Error(err))
		return nil, fmt.Errorf("failed to recover tunnel: %w", err)
	}

	return &proto.RecoverTunnelResponse{
		Success: true,
	}, nil
}
