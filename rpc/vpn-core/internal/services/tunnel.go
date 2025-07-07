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

// TunnelService реализация управления туннелями
type TunnelService struct {
	tunnels   map[string]*domain.Tunnel
	peers     map[string][]*domain.Peer
	keyGen    ports.KeyGenerator
	wgManager ports.WireGuardManager
	logger    *zap.Logger
	mutex     sync.RWMutex

	// Новые поля для мониторинга
	tunnelStartTimes map[string]time.Time
	errorCounts      map[string]int
	recoveryCounts   map[string]int
}

// GetTunnels returns the internal tunnels map for testing purposes.
func (t *TunnelService) GetTunnels() map[string]*domain.Tunnel {
	return t.tunnels
}

// GetPeers returns the internal peers map for testing purposes.
func (t *TunnelService) GetPeers() map[string][]*domain.Peer {
	return t.peers
}

// GetErrorCounts returns the internal errorCounts map for testing purposes.
func (t *TunnelService) GetErrorCounts() map[string]int {
	return t.errorCounts
}

// GetRecoveryCounts returns the internal recoveryCounts map for testing purposes.
func (t *TunnelService) GetRecoveryCounts() map[string]int {
	return t.recoveryCounts
}

// NewTunnelService создает новый сервис управления туннелями
func NewTunnelService(keyGen ports.KeyGenerator, wgManager ports.WireGuardManager, logger *zap.Logger) ports.TunnelManager {
	return &TunnelService{
		tunnels:          make(map[string]*domain.Tunnel),
		peers:            make(map[string][]*domain.Peer),
		keyGen:           keyGen,
		wgManager:        wgManager,
		logger:           logger,
		tunnelStartTimes: make(map[string]time.Time),
		errorCounts:      make(map[string]int),
		recoveryCounts:   make(map[string]int),
	}
}

// CreateTunnel создает новый туннель
func (t *TunnelService) CreateTunnel(ctx context.Context, req *domain.CreateTunnelRequest) (*domain.Tunnel, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	publicKey, privateKey, err := t.keyGen.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	tunnel := &domain.Tunnel{
		ID:               generateID(),
		Name:             req.Name,
		Interface:        fmt.Sprintf("wg%d", len(t.tunnels)),
		Status:           domain.TunnelStatusInactive,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		ListenPort:       req.ListenPort,
		MTU:              req.MTU,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		AutoRecovery:     req.AutoRecovery,
		RecoveryAttempts: 0,
	}

	t.tunnels[tunnel.ID] = tunnel
	t.peers[tunnel.ID] = []*domain.Peer{}

	t.logger.Info("tunnel created",
		zap.String("id", tunnel.ID),
		zap.String("name", tunnel.Name),
		zap.String("interface", tunnel.Interface),
		zap.Bool("auto_recovery", tunnel.AutoRecovery))

	return tunnel, nil
}

// GetTunnel получает туннель по ID
func (t *TunnelService) GetTunnel(ctx context.Context, id string) (*domain.Tunnel, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tunnel, exists := t.tunnels[id]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", id)
	}

	return tunnel, nil
}

// ListTunnels возвращает список всех туннелей
func (t *TunnelService) ListTunnels(ctx context.Context) ([]*domain.Tunnel, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tunnels := make([]*domain.Tunnel, 0, len(t.tunnels))
	for _, tunnel := range t.tunnels {
		tunnels = append(tunnels, tunnel)
	}

	return tunnels, nil
}

// DeleteTunnel удаляет туннель
func (t *TunnelService) DeleteTunnel(ctx context.Context, id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.tunnels[id]; !exists {
		return fmt.Errorf("tunnel not found: %s", id)
	}

	delete(t.tunnels, id)
	delete(t.peers, id)
	delete(t.tunnelStartTimes, id)
	delete(t.errorCounts, id)
	delete(t.recoveryCounts, id)

	t.logger.Info("tunnel deleted", zap.String("id", id))
	return nil
}

// StartTunnel запускает туннель
func (t *TunnelService) StartTunnel(ctx context.Context, id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	tunnel, exists := t.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", id)
	}

	if err := t.wgManager.CreateInterface(tunnel.Interface, tunnel.PrivateKey, tunnel.ListenPort, tunnel.MTU); err != nil {
		tunnel.Status = domain.TunnelStatusError
		tunnel.UpdatedAt = time.Now()
		t.errorCounts[id]++
		return fmt.Errorf("failed to create wireguard interface: %w", err)
	}

	tunnel.Status = domain.TunnelStatusActive
	tunnel.UpdatedAt = time.Now()
	t.tunnelStartTimes[id] = time.Now()

	t.logger.Info("tunnel started",
		zap.String("id", id),
		zap.String("interface", tunnel.Interface))
	return nil
}

// StopTunnel останавливает туннель
func (t *TunnelService) StopTunnel(ctx context.Context, id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	tunnel, exists := t.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", id)
	}

	if err := t.wgManager.DeleteInterface(tunnel.Interface); err != nil {
		tunnel.Status = domain.TunnelStatusError
		tunnel.UpdatedAt = time.Now()
		t.errorCounts[id]++
		return fmt.Errorf("failed to delete wireguard interface: %w", err)
	}

	tunnel.Status = domain.TunnelStatusInactive
	tunnel.UpdatedAt = time.Now()
	delete(t.tunnelStartTimes, id)

	t.logger.Info("tunnel stopped",
		zap.String("id", id),
		zap.String("interface", tunnel.Interface))
	return nil
}
