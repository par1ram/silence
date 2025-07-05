package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"go.uber.org/zap"
)

// GetTunnelStats получает статистику туннеля
func (t *TunnelService) GetTunnelStats(ctx context.Context, id string) (*domain.TunnelStats, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tunnel, exists := t.tunnels[id]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", id)
	}

	peers := t.peers[id]
	activePeers := 0
	for _, peer := range peers {
		if peer.Status == domain.PeerStatusActive {
			activePeers++
		}
	}

	var bytesRx, bytesTx int64
	if stats, err := t.wgManager.GetInterfaceStats(tunnel.Interface); err == nil {
		bytesRx = stats.BytesRx
		bytesTx = stats.BytesTx
	}

	var uptime time.Duration
	if startTime, exists := t.tunnelStartTimes[id]; exists {
		uptime = time.Since(startTime)
	}

	return &domain.TunnelStats{
		TunnelID:      id,
		BytesRx:       bytesRx,
		BytesTx:       bytesTx,
		PeersCount:    len(peers),
		ActivePeers:   activePeers,
		LastUpdated:   time.Now(),
		Uptime:        uptime,
		ErrorCount:    t.errorCounts[id],
		RecoveryCount: t.recoveryCounts[id],
	}, nil
}

// HealthCheck проверяет здоровье туннеля
func (t *TunnelService) HealthCheck(ctx context.Context, req *domain.HealthCheckRequest) (*domain.HealthCheckResponse, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tunnel, exists := t.tunnels[req.TunnelID]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", req.TunnelID)
	}

	status := "healthy"
	if tunnel.Status != domain.TunnelStatusActive {
		status = "unhealthy"
	}

	peersHealth := make([]domain.PeerHealth, 0)
	peers := t.peers[req.TunnelID]
	for _, peer := range peers {
		peerHealth := domain.PeerHealth{
			PeerID:            peer.ID,
			Status:            peer.Status,
			LastHandshake:     peer.LastHandshake,
			Latency:           peer.Latency,
			PacketLoss:        peer.PacketLoss,
			ConnectionQuality: peer.ConnectionQuality,
		}
		peersHealth = append(peersHealth, peerHealth)
	}

	var uptime time.Duration
	if startTime, exists := t.tunnelStartTimes[req.TunnelID]; exists {
		uptime = time.Since(startTime)
	}

	return &domain.HealthCheckResponse{
		TunnelID:    req.TunnelID,
		Status:      status,
		LastCheck:   time.Now(),
		PeersHealth: peersHealth,
		Uptime:      uptime,
		ErrorCount:  t.errorCounts[req.TunnelID],
	}, nil
}

// EnableAutoRecovery включает автоматическое восстановление для туннеля
func (t *TunnelService) EnableAutoRecovery(ctx context.Context, tunnelID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	tunnel, exists := t.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	tunnel.AutoRecovery = true
	tunnel.UpdatedAt = time.Now()

	t.logger.Info("auto recovery enabled", zap.String("tunnel_id", tunnelID))
	return nil
}

// DisableAutoRecovery отключает автоматическое восстановление для туннеля
func (t *TunnelService) DisableAutoRecovery(ctx context.Context, tunnelID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	tunnel, exists := t.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	tunnel.AutoRecovery = false
	tunnel.UpdatedAt = time.Now()

	t.logger.Info("auto recovery disabled", zap.String("tunnel_id", tunnelID))
	return nil
}

// RecoverTunnel восстанавливает туннель
func (t *TunnelService) RecoverTunnel(ctx context.Context, tunnelID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	tunnel, exists := t.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	t.logger.Info("recovering tunnel", zap.String("tunnel_id", tunnelID))

	if tunnel.Status == domain.TunnelStatusActive {
		if err := t.wgManager.DeleteInterface(tunnel.Interface); err != nil {
			t.logger.Warn("failed to delete interface during recovery",
				zap.String("tunnel_id", tunnelID),
				zap.Error(err))
		}
	}
	time.Sleep(2 * time.Second)

	if err := t.wgManager.CreateInterface(tunnel.Interface, tunnel.PrivateKey, tunnel.ListenPort, tunnel.MTU); err != nil {
		tunnel.Status = domain.TunnelStatusError
		tunnel.UpdatedAt = time.Now()
		t.errorCounts[tunnelID]++
		return fmt.Errorf("failed to recreate wireguard interface: %w", err)
	}

	tunnel.Status = domain.TunnelStatusActive
	tunnel.UpdatedAt = time.Now()
	t.tunnelStartTimes[tunnelID] = time.Now()
	t.recoveryCounts[tunnelID]++

	t.logger.Info("tunnel recovered successfully", zap.String("tunnel_id", tunnelID))
	return nil
}
