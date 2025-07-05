package grpc

import (
	"context"
	"fmt"

	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AddPeer добавляет пира к туннелю
func (s *VpnCoreService) AddPeer(ctx context.Context, req *proto.AddPeerRequest) (*proto.Peer, error) {
	s.logger.Info("adding peer", zap.String("tunnel_id", req.TunnelId), zap.String("name", req.Name))

	domainReq := &domain.AddPeerRequest{
		TunnelID:            req.TunnelId,
		Name:                req.Name,
		PublicKey:           req.PublicKey,
		AllowedIPs:          []string{req.AllowedIps},
		Endpoint:            req.Endpoint,
		PersistentKeepalive: int(req.Keepalive),
	}

	peer, err := s.peerManager.AddPeer(ctx, domainReq)
	if err != nil {
		s.logger.Error("failed to add peer", zap.Error(err))
		return nil, fmt.Errorf("failed to add peer: %w", err)
	}

	return s.domainPeerToProto(peer), nil
}

// GetPeer получает пира по ID
func (s *VpnCoreService) GetPeer(ctx context.Context, req *proto.GetPeerRequest) (*proto.Peer, error) {
	s.logger.Debug("getting peer", zap.String("tunnel_id", req.TunnelId), zap.String("peer_id", req.PeerId))

	peer, err := s.peerManager.GetPeer(ctx, req.TunnelId, req.PeerId)
	if err != nil {
		s.logger.Error("failed to get peer", zap.Error(err))
		return nil, fmt.Errorf("failed to get peer: %w", err)
	}

	return s.domainPeerToProto(peer), nil
}

// ListPeers возвращает список пиров туннеля
func (s *VpnCoreService) ListPeers(ctx context.Context, req *proto.ListPeersRequest) (*proto.ListPeersResponse, error) {
	s.logger.Debug("listing peers", zap.String("tunnel_id", req.TunnelId))

	peers, err := s.peerManager.ListPeers(ctx, req.TunnelId)
	if err != nil {
		s.logger.Error("failed to list peers", zap.Error(err))
		return nil, fmt.Errorf("failed to list peers: %w", err)
	}

	protoPeers := make([]*proto.Peer, len(peers))
	for i, peer := range peers {
		protoPeers[i] = s.domainPeerToProto(peer)
	}

	return &proto.ListPeersResponse{
		Peers: protoPeers,
	}, nil
}

// RemovePeer удаляет пира из туннеля
func (s *VpnCoreService) RemovePeer(ctx context.Context, req *proto.RemovePeerRequest) (*proto.RemovePeerResponse, error) {
	s.logger.Info("removing peer", zap.String("tunnel_id", req.TunnelId), zap.String("peer_id", req.PeerId))

	err := s.peerManager.RemovePeer(ctx, req.TunnelId, req.PeerId)
	if err != nil {
		s.logger.Error("failed to remove peer", zap.Error(err))
		return nil, fmt.Errorf("failed to remove peer: %w", err)
	}

	return &proto.RemovePeerResponse{
		Success: true,
	}, nil
}

// domainPeerToProto конвертирует доменную модель пира в proto
func (s *VpnCoreService) domainPeerToProto(peer *domain.Peer) *proto.Peer {
	allowedIPs := ""
	if len(peer.AllowedIPs) > 0 {
		allowedIPs = peer.AllowedIPs[0]
	}

	// Конвертируем статус
	var status proto.PeerStatus
	switch peer.Status {
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

	protoPeer := &proto.Peer{
		Id:         peer.ID,
		TunnelId:   peer.TunnelID,
		Name:       peer.Name,
		PublicKey:  peer.PublicKey,
		AllowedIps: allowedIPs,
		Endpoint:   peer.Endpoint,
		Keepalive:  int32(peer.PersistentKeepalive),
		Status:     status,
		CreatedAt:  timestamppb.New(peer.CreatedAt),
		UpdatedAt:  timestamppb.New(peer.UpdatedAt),
	}

	// Добавляем новые поля для мониторинга
	if !peer.LastSeen.IsZero() {
		protoPeer.LastSeen = timestamppb.New(peer.LastSeen)
	}
	if peer.ConnectionQuality > 0 {
		protoPeer.ConnectionQuality = peer.ConnectionQuality
	}
	if peer.Latency > 0 {
		protoPeer.Latency = int64(peer.Latency.Milliseconds())
	}
	if peer.PacketLoss > 0 {
		protoPeer.PacketLoss = peer.PacketLoss
	}

	return protoPeer
}
