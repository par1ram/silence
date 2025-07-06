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

func TestVpnCoreService_AddPeer(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.AddPeerRequest
		mockPeer      *domain.Peer
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное добавление пира",
			request: &proto.AddPeerRequest{
				TunnelId:   "tunnel-1",
				Name:       "test-peer",
				PublicKey:  "test-public-key",
				AllowedIps: "10.0.0.2/32",
				Endpoint:   "192.168.1.100:51820",
				Keepalive:  25,
			},
			mockPeer: &domain.Peer{
				ID:                  "peer-1",
				TunnelID:            "tunnel-1",
				Name:                "test-peer",
				PublicKey:           "test-public-key",
				AllowedIPs:          []string{"10.0.0.2/32"},
				Endpoint:            "192.168.1.100:51820",
				PersistentKeepalive: 25,
				Status:              domain.PeerStatusActive,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expectedError: false,
		},
		{
			name: "ошибка добавления пира",
			request: &proto.AddPeerRequest{
				TunnelId:   "tunnel-1",
				Name:       "test-peer",
				PublicKey:  "invalid-key",
				AllowedIps: "10.0.0.2/32",
				Endpoint:   "192.168.1.100:51820",
				Keepalive:  25,
			},
			mockError:     errors.New("invalid public key"),
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
				mockPeerManager.EXPECT().
					AddPeer(gomock.Any(), gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				mockPeerManager.EXPECT().
					AddPeer(gomock.Any(), gomock.Any()).
					Return(tt.mockPeer, nil)
			}

			result, err := service.AddPeer(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockPeer.ID, result.Id)
				assert.Equal(t, tt.mockPeer.Name, result.Name)
				assert.Equal(t, tt.mockPeer.PublicKey, result.PublicKey)
			}
		})
	}
}

func TestVpnCoreService_GetPeer(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.GetPeerRequest
		mockPeer      *domain.Peer
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное получение пира",
			request: &proto.GetPeerRequest{
				TunnelId: "tunnel-1",
				PeerId:   "peer-1",
			},
			mockPeer: &domain.Peer{
				ID:         "peer-1",
				TunnelID:   "tunnel-1",
				Name:       "test-peer",
				PublicKey:  "test-public-key",
				AllowedIPs: []string{"10.0.0.2/32"},
				Endpoint:   "192.168.1.100:51820",
				Status:     domain.PeerStatusActive,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectedError: false,
		},
		{
			name: "пир не найден",
			request: &proto.GetPeerRequest{
				TunnelId: "tunnel-1",
				PeerId:   "non-existent",
			},
			mockError:     errors.New("peer not found"),
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
				mockPeerManager.EXPECT().
					GetPeer(gomock.Any(), tt.request.TunnelId, tt.request.PeerId).
					Return(nil, tt.mockError)
			} else {
				mockPeerManager.EXPECT().
					GetPeer(gomock.Any(), tt.request.TunnelId, tt.request.PeerId).
					Return(tt.mockPeer, nil)
			}

			result, err := service.GetPeer(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockPeer.ID, result.Id)
				assert.Equal(t, tt.mockPeer.Name, result.Name)
				assert.Equal(t, tt.mockPeer.PublicKey, result.PublicKey)
			}
		})
	}
}

func TestVpnCoreService_ListPeers(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.ListPeersRequest
		mockPeers     []*domain.Peer
		mockError     error
		expectedError bool
		expectedCount int
	}{
		{
			name: "успешное получение списка пиров",
			request: &proto.ListPeersRequest{
				TunnelId: "tunnel-1",
			},
			mockPeers: []*domain.Peer{
				{
					ID:         "peer-1",
					TunnelID:   "tunnel-1",
					Name:       "test-peer-1",
					PublicKey:  "test-public-key-1",
					AllowedIPs: []string{"10.0.0.2/32"},
					Endpoint:   "192.168.1.100:51820",
					Status:     domain.PeerStatusActive,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
				{
					ID:         "peer-2",
					TunnelID:   "tunnel-1",
					Name:       "test-peer-2",
					PublicKey:  "test-public-key-2",
					AllowedIPs: []string{"10.0.0.3/32"},
					Endpoint:   "192.168.1.101:51820",
					Status:     domain.PeerStatusInactive,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name: "пустой список пиров",
			request: &proto.ListPeersRequest{
				TunnelId: "tunnel-1",
			},
			mockPeers:     []*domain.Peer{},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name: "ошибка получения списка пиров",
			request: &proto.ListPeersRequest{
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
				mockPeerManager.EXPECT().
					ListPeers(gomock.Any(), tt.request.TunnelId).
					Return(nil, tt.mockError)
			} else {
				mockPeerManager.EXPECT().
					ListPeers(gomock.Any(), tt.request.TunnelId).
					Return(tt.mockPeers, nil)
			}

			result, err := service.ListPeers(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Peers, tt.expectedCount)
			}
		})
	}
}

func TestVpnCoreService_RemovePeer(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.RemovePeerRequest
		mockError     error
		expectedError bool
	}{
		{
			name: "успешное удаление пира",
			request: &proto.RemovePeerRequest{
				TunnelId: "tunnel-1",
				PeerId:   "peer-1",
			},
			expectedError: false,
		},
		{
			name: "ошибка удаления пира",
			request: &proto.RemovePeerRequest{
				TunnelId: "tunnel-1",
				PeerId:   "peer-1",
			},
			mockError:     errors.New("peer not found"),
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
				mockPeerManager.EXPECT().
					RemovePeer(gomock.Any(), tt.request.TunnelId, tt.request.PeerId).
					Return(tt.mockError)
			} else {
				mockPeerManager.EXPECT().
					RemovePeer(gomock.Any(), tt.request.TunnelId, tt.request.PeerId).
					Return(nil)
			}

			result, err := service.RemovePeer(context.Background(), tt.request)

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

func TestVpnCoreService_domainPeerToProto(t *testing.T) {
	tests := []struct {
		name     string
		peer     *domain.Peer
		expected *proto.Peer
	}{
		{
			name: "конвертация активного пира",
			peer: &domain.Peer{
				ID:                  "peer-1",
				TunnelID:            "tunnel-1",
				Name:                "test-peer",
				PublicKey:           "test-public-key",
				AllowedIPs:          []string{"10.0.0.2/32"},
				Endpoint:            "192.168.1.100:51820",
				PersistentKeepalive: 25,
				Status:              domain.PeerStatusActive,
				LastSeen:            time.Now(),
				ConnectionQuality:   95.5,
				Latency:             10 * time.Millisecond,
				PacketLoss:          0.1,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expected: &proto.Peer{
				Id:                "peer-1",
				TunnelId:          "tunnel-1",
				Name:              "test-peer",
				PublicKey:         "test-public-key",
				AllowedIps:        "10.0.0.2/32",
				Endpoint:          "192.168.1.100:51820",
				Keepalive:         25,
				Status:            proto.PeerStatus_PEER_STATUS_ACTIVE,
				ConnectionQuality: 95.5,
				Latency:           10,
				PacketLoss:        0.1,
			},
		},
		{
			name: "конвертация неактивного пира",
			peer: &domain.Peer{
				ID:         "peer-2",
				TunnelID:   "tunnel-1",
				Name:       "test-peer-2",
				PublicKey:  "test-public-key-2",
				AllowedIPs: []string{"10.0.0.3/32"},
				Endpoint:   "192.168.1.101:51820",
				Status:     domain.PeerStatusInactive,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &proto.Peer{
				Id:         "peer-2",
				TunnelId:   "tunnel-1",
				Name:       "test-peer-2",
				PublicKey:  "test-public-key-2",
				AllowedIps: "10.0.0.3/32",
				Endpoint:   "192.168.1.101:51820",
				Status:     proto.PeerStatus_PEER_STATUS_INACTIVE,
			},
		},
		{
			name: "конвертация пира с ошибкой",
			peer: &domain.Peer{
				ID:         "peer-3",
				TunnelID:   "tunnel-1",
				Name:       "test-peer-3",
				PublicKey:  "test-public-key-3",
				AllowedIPs: []string{"10.0.0.4/32"},
				Endpoint:   "192.168.1.102:51820",
				Status:     domain.PeerStatusError,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &proto.Peer{
				Id:         "peer-3",
				TunnelId:   "tunnel-1",
				Name:       "test-peer-3",
				PublicKey:  "test-public-key-3",
				AllowedIps: "10.0.0.4/32",
				Endpoint:   "192.168.1.102:51820",
				Status:     proto.PeerStatus_PEER_STATUS_ERROR,
			},
		},
		{
			name: "конвертация офлайн пира",
			peer: &domain.Peer{
				ID:         "peer-4",
				TunnelID:   "tunnel-1",
				Name:       "test-peer-4",
				PublicKey:  "test-public-key-4",
				AllowedIPs: []string{"10.0.0.5/32"},
				Endpoint:   "192.168.1.103:51820",
				Status:     domain.PeerStatusOffline,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expected: &proto.Peer{
				Id:         "peer-4",
				TunnelId:   "tunnel-1",
				Name:       "test-peer-4",
				PublicKey:  "test-public-key-4",
				AllowedIps: "10.0.0.5/32",
				Endpoint:   "192.168.1.103:51820",
				Status:     proto.PeerStatus_PEER_STATUS_OFFLINE,
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

			result := service.domainPeerToProto(tt.peer)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.TunnelId, result.TunnelId)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.PublicKey, result.PublicKey)
			assert.Equal(t, tt.expected.AllowedIps, result.AllowedIps)
			assert.Equal(t, tt.expected.Endpoint, result.Endpoint)
			assert.Equal(t, tt.expected.Keepalive, result.Keepalive)
			assert.Equal(t, tt.expected.Status, result.Status)

			if tt.peer.ConnectionQuality > 0 {
				assert.Equal(t, tt.expected.ConnectionQuality, result.ConnectionQuality)
			}
			if tt.peer.Latency > 0 {
				assert.Equal(t, tt.expected.Latency, result.Latency)
			}
			if tt.peer.PacketLoss > 0 {
				assert.Equal(t, tt.expected.PacketLoss, result.PacketLoss)
			}
		})
	}
}
