package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	mocks "github.com/par1ram/silence/rpc/vpn-core/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestVpnCoreService_CreateTunnel(t *testing.T) {
	tests := []struct {
		name           string
		request        *proto.CreateTunnelRequest
		mockTunnel     *domain.Tunnel
		mockError      error
		expectedError  bool
		expectedTunnel *proto.Tunnel
	}{
		{
			name: "успешное создание туннеля",
			request: &proto.CreateTunnelRequest{
				Name:         "test-tunnel",
				ListenPort:   51820,
				Mtu:          1420,
				AutoRecovery: true,
			},
			mockTunnel: &domain.Tunnel{
				ID:           "tunnel-1",
				Name:         "test-tunnel",
				ListenPort:   51820,
				MTU:          1420,
				Status:       domain.TunnelStatusInactive,
				AutoRecovery: true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			expectedError: false,
		},
		{
			name: "ошибка создания туннеля",
			request: &proto.CreateTunnelRequest{
				Name:         "test-tunnel",
				ListenPort:   51820,
				Mtu:          1420,
				AutoRecovery: true,
			},
			mockError:     errors.New("port already in use"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					CreateTunnel(gomock.Any(), gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					CreateTunnel(gomock.Any(), gomock.Any()).
					Return(tt.mockTunnel, nil)
			}

			result, err := service.CreateTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockTunnel.ID, result.Id)
				assert.Equal(t, tt.mockTunnel.Name, result.Name)
			}
		})
	}
}

func TestVpnCoreService_GetTunnel(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.GetTunnelRequest
		mockTunnel    *domain.Tunnel
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное получение туннеля",
			request: &proto.GetTunnelRequest{
				Id: "tunnel-1",
			},
			mockTunnel: &domain.Tunnel{
				ID:           "tunnel-1",
				Name:         "test-tunnel",
				ListenPort:   51820,
				MTU:          1420,
				Status:       domain.TunnelStatusActive,
				AutoRecovery: true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			expectedError: false,
		},
		{
			name: "туннель не найден",
			request: &proto.GetTunnelRequest{
				Id: "non-existent",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					GetTunnel(gomock.Any(), tt.request.Id).
					Return(nil, tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					GetTunnel(gomock.Any(), tt.request.Id).
					Return(tt.mockTunnel, nil)
			}

			result, err := service.GetTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockTunnel.ID, result.Id)
				assert.Equal(t, tt.mockTunnel.Name, result.Name)
			}
		})
	}
}

func TestVpnCoreService_ListTunnels(t *testing.T) {
	tests := []struct {
		name          string
		mockTunnels   []*domain.Tunnel
		mockError     error
		expectedError bool
		expectedCount int
	}{
		{
			name: "успешное получение списка туннелей",
			mockTunnels: []*domain.Tunnel{
				{
					ID:           "tunnel-1",
					Name:         "test-tunnel-1",
					ListenPort:   51820,
					MTU:          1420,
					Status:       domain.TunnelStatusActive,
					AutoRecovery: true,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
				{
					ID:           "tunnel-2",
					Name:         "test-tunnel-2",
					ListenPort:   51821,
					MTU:          1420,
					Status:       domain.TunnelStatusInactive,
					AutoRecovery: false,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "пустой список туннелей",
			mockTunnels:   []*domain.Tunnel{},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "ошибка получения списка",
			mockError:     errors.New("database error"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					ListTunnels(gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					ListTunnels(gomock.Any()).
					Return(tt.mockTunnels, nil)
			}

			result, err := service.ListTunnels(context.Background(), &proto.ListTunnelsRequest{})

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Tunnels, tt.expectedCount)
			}
		})
	}
}

func TestVpnCoreService_DeleteTunnel(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.DeleteTunnelRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное удаление туннеля",
			request: &proto.DeleteTunnelRequest{
				Id: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка удаления туннеля",
			request: &proto.DeleteTunnelRequest{
				Id: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					DeleteTunnel(gomock.Any(), tt.request.Id).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					DeleteTunnel(gomock.Any(), tt.request.Id).
					Return(nil)
			}

			result, err := service.DeleteTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestVpnCoreService_StartTunnel(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.StartTunnelRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешный запуск туннеля",
			request: &proto.StartTunnelRequest{
				Id: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка запуска туннеля",
			request: &proto.StartTunnelRequest{
				Id: "tunnel-1",
			},
			mockError:     errors.New("port already in use"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					StartTunnel(gomock.Any(), tt.request.Id).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					StartTunnel(gomock.Any(), tt.request.Id).
					Return(nil)
			}

			result, err := service.StartTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestVpnCoreService_StopTunnel(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.StopTunnelRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешная остановка туннеля",
			request: &proto.StopTunnelRequest{
				Id: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка остановки туннеля",
			request: &proto.StopTunnelRequest{
				Id: "tunnel-1",
			},
			mockError:     errors.New("tunnel not running"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					StopTunnel(gomock.Any(), tt.request.Id).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					StopTunnel(gomock.Any(), tt.request.Id).
					Return(nil)
			}

			result, err := service.StopTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestVpnCoreService_GetTunnelStats(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.GetTunnelStatsRequest
		mockStats     *domain.TunnelStats
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное получение статистики",
			request: &proto.GetTunnelStatsRequest{
				Id: "tunnel-1",
			},
			mockStats: &domain.TunnelStats{
				TunnelID:      "tunnel-1",
				BytesRx:       1024,
				BytesTx:       2048,
				PeersCount:    5,
				ActivePeers:   3,
				LastUpdated:   time.Now(),
				Uptime:        3600 * time.Second,
				ErrorCount:    2,
				RecoveryCount: 1,
			},
			expectedError: false,
		},
		{
			name: "ошибка получения статистики",
			request: &proto.GetTunnelStatsRequest{
				Id: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					GetTunnelStats(gomock.Any(), tt.request.Id).
					Return(nil, tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					GetTunnelStats(gomock.Any(), tt.request.Id).
					Return(tt.mockStats, nil)
			}

			result, err := service.GetTunnelStats(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockStats.TunnelID, result.TunnelId)
				assert.Equal(t, tt.mockStats.BytesRx, result.BytesRx)
				assert.Equal(t, tt.mockStats.BytesTx, result.BytesTx)
				assert.Equal(t, int32(tt.mockStats.PeersCount), result.PeersCount)
				assert.Equal(t, int32(tt.mockStats.ActivePeers), result.ActivePeers)
			}
		})
	}
}

func TestVpnCoreService_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.HealthCheckRequest
		mockHealth    *domain.HealthCheckResponse
		mockError     error
		expectedError bool
	}{
		{
			name: "успешная проверка здоровья",
			request: &proto.HealthCheckRequest{
				TunnelId: "tunnel-1",
			},
			mockHealth: &domain.HealthCheckResponse{
				TunnelID:   "tunnel-1",
				Status:     "healthy",
				LastCheck:  time.Now(),
				Uptime:     3600 * time.Second,
				ErrorCount: 0,
				PeersHealth: []domain.PeerHealth{
					{
						PeerID:            "peer-1",
						Status:            domain.PeerStatusActive,
						LastHandshake:     time.Now(),
						Latency:           10 * time.Millisecond,
						PacketLoss:        0.1,
						ConnectionQuality: 95,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "ошибка проверки здоровья",
			request: &proto.HealthCheckRequest{
				TunnelId: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					HealthCheck(gomock.Any(), gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					HealthCheck(gomock.Any(), gomock.Any()).
					Return(tt.mockHealth, nil)
			}

			result, err := service.HealthCheck(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockHealth.TunnelID, result.TunnelId)
				assert.Equal(t, tt.mockHealth.Status, result.Status)
				assert.Len(t, result.PeersHealth, len(tt.mockHealth.PeersHealth))
			}
		})
	}
}

func TestVpnCoreService_EnableAutoRecovery(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.EnableAutoRecoveryRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное включение авто-восстановления",
			request: &proto.EnableAutoRecoveryRequest{
				TunnelId: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка включения авто-восстановления",
			request: &proto.EnableAutoRecoveryRequest{
				TunnelId: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					EnableAutoRecovery(gomock.Any(), tt.request.TunnelId).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					EnableAutoRecovery(gomock.Any(), tt.request.TunnelId).
					Return(nil)
			}

			result, err := service.EnableAutoRecovery(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestVpnCoreService_DisableAutoRecovery(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.DisableAutoRecoveryRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное отключение авто-восстановления",
			request: &proto.DisableAutoRecoveryRequest{
				TunnelId: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка отключения авто-восстановления",
			request: &proto.DisableAutoRecoveryRequest{
				TunnelId: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					DisableAutoRecovery(gomock.Any(), tt.request.TunnelId).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					DisableAutoRecovery(gomock.Any(), tt.request.TunnelId).
					Return(nil)
			}

			result, err := service.DisableAutoRecovery(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}

func TestVpnCoreService_RecoverTunnel(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.RecoverTunnelRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное восстановление туннеля",
			request: &proto.RecoverTunnelRequest{
				TunnelId: "tunnel-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка восстановления туннеля",
			request: &proto.RecoverTunnelRequest{
				TunnelId: "tunnel-1",
			},
			mockError:     errors.New("tunnel not found"),
			expectedError: true,
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

			if tt.mockError != nil {
				mockTunnelManager.EXPECT().
					RecoverTunnel(gomock.Any(), tt.request.TunnelId).
					Return(tt.mockError)
			} else {
				mockTunnelManager.EXPECT().
					RecoverTunnel(gomock.Any(), tt.request.TunnelId).
					Return(nil)
			}

			result, err := service.RecoverTunnel(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
			}
		})
	}
}
