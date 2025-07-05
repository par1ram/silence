package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
)

// PeerService реализация управления пирами
type PeerService struct {
	peers  map[string]map[string]*domain.Peer // tunnelID -> peerID -> peer
	logger *zap.Logger
	mutex  sync.RWMutex
}

// NewPeerService создает новый сервис управления пирами
func NewPeerService(logger *zap.Logger) ports.PeerManager {
	return &PeerService{
		peers:  make(map[string]map[string]*domain.Peer),
		logger: logger,
	}
}

// AddPeer добавляет пира в туннель
func (p *PeerService) AddPeer(ctx context.Context, req *domain.AddPeerRequest) (*domain.Peer, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Инициализируем map для туннеля, если не существует
	if p.peers[req.TunnelID] == nil {
		p.peers[req.TunnelID] = make(map[string]*domain.Peer)
	}

	peerID := generatePeerID()
	peer := &domain.Peer{
		ID:                  peerID,
		TunnelID:            req.TunnelID,
		Name:                req.Name,
		PublicKey:           req.PublicKey,
		AllowedIPs:          req.AllowedIPs,
		Endpoint:            req.Endpoint,
		PersistentKeepalive: req.PersistentKeepalive,
		Status:              domain.PeerStatusInactive,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	p.peers[req.TunnelID][peerID] = peer

	p.logger.Info("peer added",
		zap.String("peer_id", peerID),
		zap.String("tunnel_id", req.TunnelID),
		zap.String("name", req.Name),
		zap.String("public_key", req.PublicKey))

	return peer, nil
}

// GetPeer получает пира по ID
func (p *PeerService) GetPeer(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	peer, exists := tunnelPeers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	return peer, nil
}

// ListPeers возвращает список пиров туннеля
func (p *PeerService) ListPeers(ctx context.Context, tunnelID string) ([]*domain.Peer, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	peers := make([]*domain.Peer, 0, len(tunnelPeers))
	for _, peer := range tunnelPeers {
		peers = append(peers, peer)
	}

	return peers, nil
}

// RemovePeer удаляет пира из туннеля
func (p *PeerService) RemovePeer(ctx context.Context, tunnelID, peerID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	if _, exists := tunnelPeers[peerID]; !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	delete(tunnelPeers, peerID)

	p.logger.Info("peer removed",
		zap.String("peer_id", peerID),
		zap.String("tunnel_id", tunnelID))

	return nil
}
