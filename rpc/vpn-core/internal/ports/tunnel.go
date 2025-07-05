package ports

import (
	"context"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
)

// TunnelManager интерфейс для управления туннелями
type TunnelManager interface {
	CreateTunnel(ctx context.Context, req *domain.CreateTunnelRequest) (*domain.Tunnel, error)
	GetTunnel(ctx context.Context, id string) (*domain.Tunnel, error)
	ListTunnels(ctx context.Context) ([]*domain.Tunnel, error)
	DeleteTunnel(ctx context.Context, id string) error
	StartTunnel(ctx context.Context, id string) error
	StopTunnel(ctx context.Context, id string) error
	GetTunnelStats(ctx context.Context, id string) (*domain.TunnelStats, error)
	// Новые методы для мониторинга и восстановления
	HealthCheck(ctx context.Context, req *domain.HealthCheckRequest) (*domain.HealthCheckResponse, error)
	EnableAutoRecovery(ctx context.Context, tunnelID string) error
	DisableAutoRecovery(ctx context.Context, tunnelID string) error
	RecoverTunnel(ctx context.Context, tunnelID string) error
}

// PeerManager интерфейс для управления пирами
type PeerManager interface {
	AddPeer(ctx context.Context, req *domain.AddPeerRequest) (*domain.Peer, error)
	GetPeer(ctx context.Context, tunnelID, peerID string) (*domain.Peer, error)
	ListPeers(ctx context.Context, tunnelID string) ([]*domain.Peer, error)
	RemovePeer(ctx context.Context, tunnelID, peerID string) error
	// Новые методы для мониторинга пиров
	UpdatePeerStats(ctx context.Context, tunnelID, peerID string, stats *PeerStats) error
	GetPeerHealth(ctx context.Context, tunnelID, peerID string) (*domain.PeerHealth, error)
	EnablePeer(ctx context.Context, tunnelID, peerID string) error
	DisablePeer(ctx context.Context, tunnelID, peerID string) error
}

// KeyGenerator интерфейс для генерации ключей
type KeyGenerator interface {
	GenerateKeyPair() (publicKey, privateKey string, err error)
	ValidatePublicKey(publicKey string) bool
}

// MonitorService интерфейс для мониторинга
type MonitorService interface {
	StartMonitoring(ctx context.Context) error
	StopMonitoring(ctx context.Context) error
	GetMonitoringStatus() bool
}
