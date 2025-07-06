package services

import (
	"context"
	"errors"
	"testing"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockTunnelManager struct {
	ports.TunnelManager
	ListTunnelsFunc   func(ctx context.Context) ([]*domain.Tunnel, error)
	HealthCheckFunc   func(ctx context.Context, req *domain.HealthCheckRequest) (*domain.HealthCheckResponse, error)
	RecoverTunnelFunc func(ctx context.Context, tunnelID string) error
}

func (m *mockTunnelManager) ListTunnels(ctx context.Context) ([]*domain.Tunnel, error) {
	return m.ListTunnelsFunc(ctx)
}
func (m *mockTunnelManager) HealthCheck(ctx context.Context, req *domain.HealthCheckRequest) (*domain.HealthCheckResponse, error) {
	return m.HealthCheckFunc(ctx, req)
}
func (m *mockTunnelManager) RecoverTunnel(ctx context.Context, tunnelID string) error {
	return m.RecoverTunnelFunc(ctx, tunnelID)
}

type mockPeerManager struct {
	ports.PeerManager
	GetPeerFunc func(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error)
}

func (m *mockPeerManager) GetPeer(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error) {
	return m.GetPeerFunc(ctx, tunnelID, peerID)
}

func TestMonitorService_performHealthChecks(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	activeTunnel := &domain.Tunnel{ID: "1", Status: domain.TunnelStatusActive, AutoRecovery: true}
	inactiveTunnel := &domain.Tunnel{ID: "2", Status: domain.TunnelStatusInactive}

	tm := &mockTunnelManager{
		ListTunnelsFunc: func(ctx context.Context) ([]*domain.Tunnel, error) {
			return []*domain.Tunnel{activeTunnel, inactiveTunnel}, nil
		},
		HealthCheckFunc: func(ctx context.Context, req *domain.HealthCheckRequest) (*domain.HealthCheckResponse, error) {
			if req.TunnelID == "1" {
				return &domain.HealthCheckResponse{TunnelID: "1", Status: "healthy"}, nil
			}
			return nil, errors.New("fail")
		},
		RecoverTunnelFunc: func(ctx context.Context, tunnelID string) error {
			return nil
		},
	}
	pm := &mockPeerManager{
		GetPeerFunc: func(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error) {
			return &domain.Peer{ID: peerID, TunnelID: tunnelID}, nil
		},
	}
	wgMock := &mockWGManager{}

	ms := NewMonitorService(tm, pm, wgMock, logger).(*MonitorService)
	ms.maxRecoveryAttempts = 2

	ms.performHealthChecks(ctx)
}

func TestMonitorService_attemptTunnelRecovery(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	tun := &domain.Tunnel{ID: "1", Status: domain.TunnelStatusActive, AutoRecovery: true}

	tm := &mockTunnelManager{
		RecoverTunnelFunc: func(ctx context.Context, tunnelID string) error {
			if tun.RecoveryAttempts == 1 { // после первого вызова RecoveryAttempts = 1
				return errors.New("fail")
			}
			return nil
		},
	}
	ms := NewMonitorService(tm, nil, nil, logger).(*MonitorService)
	ms.maxRecoveryAttempts = 2

	// Первый вызов - ошибка восстановления
	ms.attemptTunnelRecovery(ctx, tun)
	assert.Equal(t, domain.TunnelStatusRecovering, tun.Status)
	assert.Equal(t, 1, tun.RecoveryAttempts)

	// Второй вызов - успешное восстановление
	ms.attemptTunnelRecovery(ctx, tun)
	assert.Equal(t, domain.TunnelStatusActive, tun.Status)
	assert.Equal(t, 0, tun.RecoveryAttempts)
}

func TestMonitorService_updateTunnelStats(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	tun := &domain.Tunnel{ID: "1", Interface: "wg0"}
	wgMock := &mockWGManager{GetInterfaceStatsFunc: func(string) (*ports.InterfaceStats, error) {
		return &ports.InterfaceStats{BytesRx: 100, BytesTx: 200, PeersCount: 2}, nil
	}}
	ms := NewMonitorService(nil, nil, wgMock, logger).(*MonitorService)
	health := &domain.HealthCheckResponse{Status: "healthy"}
	ms.updateTunnelStats(ctx, tun, health)
	assert.Equal(t, "healthy", tun.HealthStatus)
}

func TestMonitorService_checkPeersHealth(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	tun := &domain.Tunnel{ID: "1"}
	pm := &mockPeerManager{
		GetPeerFunc: func(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error) {
			return &domain.Peer{ID: peerID, TunnelID: tunnelID}, nil
		},
	}
	ms := NewMonitorService(nil, pm, nil, logger).(*MonitorService)
	peersHealth := []domain.PeerHealth{{PeerID: "p1"}, {PeerID: "p2"}}
	ms.checkPeersHealth(ctx, tun, peersHealth)
}
