package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
)

// UpdatePeerStats обновляет статистику пира
func (p *PeerService) UpdatePeerStats(ctx context.Context, tunnelID, peerID string, stats *ports.PeerStats) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	peer, exists := tunnelPeers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.TransferRx = stats.TransferRx
	peer.TransferTx = stats.TransferTx
	peer.LastHandshake = time.Unix(stats.LastHandshake, 0)
	peer.LastSeen = time.Now()
	peer.UpdatedAt = time.Now()

	if stats.LastHandshake > 0 {
		timeSinceHandshake := time.Since(peer.LastHandshake)
		if timeSinceHandshake < 2*time.Minute {
			peer.Status = domain.PeerStatusActive
		} else if timeSinceHandshake < 10*time.Minute {
			peer.Status = domain.PeerStatusInactive
		} else {
			peer.Status = domain.PeerStatusOffline
		}
	}

	p.logger.Debug("peer stats updated",
		zap.String("peer_id", peerID),
		zap.String("tunnel_id", tunnelID),
		zap.String("status", string(peer.Status)),
		zap.Int64("transfer_rx", stats.TransferRx),
		zap.Int64("transfer_tx", stats.TransferTx))

	return nil
}

// GetPeerHealth получает здоровье пира
func (p *PeerService) GetPeerHealth(ctx context.Context, tunnelID, peerID string) (*domain.PeerHealth, error) {
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

	connectionQuality := p.calculateConnectionQuality(peer)

	peerHealth := &domain.PeerHealth{
		PeerID:            peer.ID,
		Status:            peer.Status,
		LastHandshake:     peer.LastHandshake,
		Latency:           peer.Latency,
		PacketLoss:        peer.PacketLoss,
		ConnectionQuality: connectionQuality,
	}

	return peerHealth, nil
}

// EnablePeer активирует пира
func (p *PeerService) EnablePeer(ctx context.Context, tunnelID, peerID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	peer, exists := tunnelPeers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.Status = domain.PeerStatusActive
	peer.UpdatedAt = time.Now()

	p.logger.Info("peer enabled",
		zap.String("peer_id", peerID),
		zap.String("tunnel_id", tunnelID))

	return nil
}

// DisablePeer деактивирует пира
func (p *PeerService) DisablePeer(ctx context.Context, tunnelID, peerID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	tunnelPeers, exists := p.peers[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	peer, exists := tunnelPeers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.Status = domain.PeerStatusInactive
	peer.UpdatedAt = time.Now()

	p.logger.Info("peer disabled",
		zap.String("peer_id", peerID),
		zap.String("tunnel_id", tunnelID))

	return nil
}

// calculateConnectionQuality вычисляет качество соединения пира
func (p *PeerService) calculateConnectionQuality(peer *domain.Peer) float64 {
	quality := 1.0

	if !peer.LastHandshake.IsZero() {
		timeSinceHandshake := time.Since(peer.LastHandshake)
		if timeSinceHandshake > 10*time.Minute {
			quality *= 0.1
		} else if timeSinceHandshake > 5*time.Minute {
			quality *= 0.3
		} else if timeSinceHandshake > 2*time.Minute {
			quality *= 0.7
		}
	}

	if peer.PacketLoss > 0 {
		quality *= (1.0 - peer.PacketLoss)
	}

	if peer.Latency > 0 {
		if peer.Latency > 500*time.Millisecond {
			quality *= 0.5
		} else if peer.Latency > 200*time.Millisecond {
			quality *= 0.8
		}
	}

	switch peer.Status {
	case domain.PeerStatusActive:
		quality *= 1.0
	case domain.PeerStatusInactive:
		quality *= 0.5
	case domain.PeerStatusOffline, domain.PeerStatusError:
		quality *= 0.1
	}

	return quality
}
