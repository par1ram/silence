package grpc

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	mocks "github.com/par1ram/silence/rpc/vpn-core/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestVpnCoreService_domainTunnelToProto(t *testing.T) {
	tests := []struct {
		name     string
		tunnel   *domain.Tunnel
		expected *proto.Tunnel
	}{
		{
			name: "конвертация активного туннеля",
			tunnel: &domain.Tunnel{
				ID:               "tunnel-1",
				Name:             "test-tunnel",
				Interface:        "wg0",
				Status:           domain.TunnelStatusActive,
				PublicKey:        "test-public-key",
				PrivateKey:       "test-private-key",
				ListenPort:       51820,
				MTU:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 2,
				LastHealthCheck:  time.Now(),
				HealthStatus:     "healthy",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expected: &proto.Tunnel{
				Id:               "tunnel-1",
				Name:             "test-tunnel",
				Interface:        "wg0",
				Status:           proto.TunnelStatus_TUNNEL_STATUS_ACTIVE,
				PublicKey:        "test-public-key",
				PrivateKey:       "test-private-key",
				ListenPort:       51820,
				Mtu:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 2,
				HealthStatus:     "healthy",
			},
		},
		{
			name: "конвертация неактивного туннеля",
			tunnel: &domain.Tunnel{
				ID:               "tunnel-2",
				Name:             "test-tunnel-2",
				Interface:        "wg1",
				Status:           domain.TunnelStatusInactive,
				PublicKey:        "test-public-key-2",
				PrivateKey:       "test-private-key-2",
				ListenPort:       51821,
				MTU:              1420,
				AutoRecovery:     false,
				RecoveryAttempts: 0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expected: &proto.Tunnel{
				Id:               "tunnel-2",
				Name:             "test-tunnel-2",
				Interface:        "wg1",
				Status:           proto.TunnelStatus_TUNNEL_STATUS_INACTIVE,
				PublicKey:        "test-public-key-2",
				PrivateKey:       "test-private-key-2",
				ListenPort:       51821,
				Mtu:              1420,
				AutoRecovery:     false,
				RecoveryAttempts: 0,
			},
		},
		{
			name: "конвертация туннеля с ошибкой",
			tunnel: &domain.Tunnel{
				ID:               "tunnel-3",
				Name:             "test-tunnel-3",
				Interface:        "wg2",
				Status:           domain.TunnelStatusError,
				PublicKey:        "test-public-key-3",
				PrivateKey:       "test-private-key-3",
				ListenPort:       51822,
				MTU:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 5,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expected: &proto.Tunnel{
				Id:               "tunnel-3",
				Name:             "test-tunnel-3",
				Interface:        "wg2",
				Status:           proto.TunnelStatus_TUNNEL_STATUS_ERROR,
				PublicKey:        "test-public-key-3",
				PrivateKey:       "test-private-key-3",
				ListenPort:       51822,
				Mtu:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 5,
			},
		},
		{
			name: "конвертация восстанавливающегося туннеля",
			tunnel: &domain.Tunnel{
				ID:               "tunnel-4",
				Name:             "test-tunnel-4",
				Interface:        "wg3",
				Status:           domain.TunnelStatusRecovering,
				PublicKey:        "test-public-key-4",
				PrivateKey:       "test-private-key-4",
				ListenPort:       51823,
				MTU:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 1,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expected: &proto.Tunnel{
				Id:               "tunnel-4",
				Name:             "test-tunnel-4",
				Interface:        "wg3",
				Status:           proto.TunnelStatus_TUNNEL_STATUS_RECOVERING,
				PublicKey:        "test-public-key-4",
				PrivateKey:       "test-private-key-4",
				ListenPort:       51823,
				Mtu:              1420,
				AutoRecovery:     true,
				RecoveryAttempts: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTunnelManager := mocks.NewMockTunnelManager(ctrl)
			mockPeerManager := mocks.NewMockPeerManager(ctrl)
			logger := zap.NewNop()

			service := NewVpnCoreService(mockTunnelManager, mockPeerManager, logger)

			result := service.domainTunnelToProto(tt.tunnel)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Interface, result.Interface)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.PublicKey, result.PublicKey)
			assert.Equal(t, tt.expected.PrivateKey, result.PrivateKey)
			assert.Equal(t, tt.expected.ListenPort, result.ListenPort)
			assert.Equal(t, tt.expected.Mtu, result.Mtu)
			assert.Equal(t, tt.expected.AutoRecovery, result.AutoRecovery)
			assert.Equal(t, tt.expected.RecoveryAttempts, result.RecoveryAttempts)

			if tt.tunnel.HealthStatus != "" {
				assert.Equal(t, tt.expected.HealthStatus, result.HealthStatus)
			}
			if !tt.tunnel.LastHealthCheck.IsZero() {
				assert.NotNil(t, result.LastHealthCheck)
			}
		})
	}
}

func TestVpnCoreService_domainTunnelStatusToProto(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.TunnelStatus
		expected proto.TunnelStatus
	}{
		{
			name:     "конвертация активного статуса",
			status:   domain.TunnelStatusActive,
			expected: proto.TunnelStatus_TUNNEL_STATUS_ACTIVE,
		},
		{
			name:     "конвертация неактивного статуса",
			status:   domain.TunnelStatusInactive,
			expected: proto.TunnelStatus_TUNNEL_STATUS_INACTIVE,
		},
		{
			name:     "конвертация статуса ошибки",
			status:   domain.TunnelStatusError,
			expected: proto.TunnelStatus_TUNNEL_STATUS_ERROR,
		},
		{
			name:     "конвертация статуса восстановления",
			status:   domain.TunnelStatusRecovering,
			expected: proto.TunnelStatus_TUNNEL_STATUS_RECOVERING,
		},
		{
			name:     "конвертация неизвестного статуса",
			status:   "unknown",
			expected: proto.TunnelStatus_TUNNEL_STATUS_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTunnelManager := mocks.NewMockTunnelManager(ctrl)
			mockPeerManager := mocks.NewMockPeerManager(ctrl)
			logger := zap.NewNop()

			service := NewVpnCoreService(mockTunnelManager, mockPeerManager, logger)

			result := service.domainTunnelStatusToProto(tt.status)

			assert.Equal(t, tt.expected, result)
		})
	}
}
