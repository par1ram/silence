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

// MonitorService реализация мониторинга туннелей
type MonitorService struct {
	tunnelManager ports.TunnelManager
	peerManager   ports.PeerManager
	wgManager     ports.WireGuardManager
	logger        *zap.Logger
	mutex         sync.RWMutex

	// Состояние мониторинга
	isMonitoring bool
	stopChan     chan struct{}

	// Конфигурация
	healthCheckInterval time.Duration
	recoveryTimeout     time.Duration
	maxRecoveryAttempts int
}

// NewMonitorService создает новый сервис мониторинга
func NewMonitorService(
	tunnelManager ports.TunnelManager,
	peerManager ports.PeerManager,
	wgManager ports.WireGuardManager,
	logger *zap.Logger,
) ports.MonitorService {
	return &MonitorService{
		tunnelManager:       tunnelManager,
		peerManager:         peerManager,
		wgManager:           wgManager,
		logger:              logger,
		healthCheckInterval: 30 * time.Second,
		recoveryTimeout:     5 * time.Minute,
		maxRecoveryAttempts: 3,
	}
}

// StartMonitoring запускает мониторинг всех туннелей
func (m *MonitorService) StartMonitoring(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isMonitoring {
		return fmt.Errorf("monitoring is already running")
	}

	m.isMonitoring = true
	m.stopChan = make(chan struct{})

	go m.monitoringLoop(ctx)

	m.logger.Info("monitoring started")
	return nil
}

// StopMonitoring останавливает мониторинг
func (m *MonitorService) StopMonitoring(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isMonitoring {
		return fmt.Errorf("monitoring is not running")
	}

	close(m.stopChan)
	m.isMonitoring = false

	m.logger.Info("monitoring stopped")
	return nil
}

// GetMonitoringStatus возвращает статус мониторинга
func (m *MonitorService) GetMonitoringStatus() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isMonitoring
}

// monitoringLoop основной цикл мониторинга
func (m *MonitorService) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(m.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks выполняет проверки здоровья всех туннелей
func (m *MonitorService) performHealthChecks(ctx context.Context) {
	tunnels, err := m.tunnelManager.ListTunnels(ctx)
	if err != nil {
		m.logger.Error("failed to list tunnels for health check", zap.Error(err))
		return
	}

	for _, tunnel := range tunnels {
		if tunnel.Status == domain.TunnelStatusActive {
			m.checkTunnelHealth(ctx, tunnel)
		}
	}
}

// checkTunnelHealth проверяет здоровье конкретного туннеля
func (m *MonitorService) checkTunnelHealth(ctx context.Context, tunnel *domain.Tunnel) {
	req := &domain.HealthCheckRequest{TunnelID: tunnel.ID}
	health, err := m.tunnelManager.HealthCheck(ctx, req)
	if err != nil {
		m.logger.Error("health check failed",
			zap.String("tunnel_id", tunnel.ID),
			zap.Error(err))

		// Если туннель настроен на автоматическое восстановление
		if tunnel.AutoRecovery && tunnel.RecoveryAttempts < m.maxRecoveryAttempts {
			m.attemptTunnelRecovery(ctx, tunnel)
		}
		return
	}

	// Обновляем статистику туннеля
	m.updateTunnelStats(ctx, tunnel, health)

	// Проверяем здоровье пиров
	m.checkPeersHealth(ctx, tunnel, health.PeersHealth)
}

// attemptTunnelRecovery пытается восстановить туннель
func (m *MonitorService) attemptTunnelRecovery(ctx context.Context, tunnel *domain.Tunnel) {
	m.logger.Info("attempting tunnel recovery",
		zap.String("tunnel_id", tunnel.ID),
		zap.Int("attempt", tunnel.RecoveryAttempts+1))

	// Устанавливаем статус восстановления
	tunnel.Status = domain.TunnelStatusRecovering
	tunnel.RecoveryAttempts++
	tunnel.UpdatedAt = time.Now()

	// Пытаемся перезапустить туннель
	if err := m.tunnelManager.RecoverTunnel(ctx, tunnel.ID); err != nil {
		m.logger.Error("tunnel recovery failed",
			zap.String("tunnel_id", tunnel.ID),
			zap.Error(err))

		// Если превышено количество попыток, устанавливаем статус ошибки
		if tunnel.RecoveryAttempts >= m.maxRecoveryAttempts {
			tunnel.Status = domain.TunnelStatusError
			tunnel.UpdatedAt = time.Now()
			m.logger.Error("tunnel recovery failed after max attempts",
				zap.String("tunnel_id", tunnel.ID),
				zap.Int("attempts", tunnel.RecoveryAttempts))
		}
	} else {
		// Восстановление успешно
		tunnel.Status = domain.TunnelStatusActive
		tunnel.RecoveryAttempts = 0
		tunnel.UpdatedAt = time.Now()
		m.logger.Info("tunnel recovery successful",
			zap.String("tunnel_id", tunnel.ID))
	}
}

// updateTunnelStats обновляет статистику туннеля
func (m *MonitorService) updateTunnelStats(ctx context.Context, tunnel *domain.Tunnel, health *domain.HealthCheckResponse) {
	tunnel.LastHealthCheck = time.Now()
	tunnel.HealthStatus = health.Status

	// Обновляем статистику через WireGuard менеджер
	if stats, err := m.wgManager.GetInterfaceStats(tunnel.Interface); err == nil {
		m.logger.Debug("updated tunnel stats",
			zap.String("tunnel_id", tunnel.ID),
			zap.Int64("bytes_rx", stats.BytesRx),
			zap.Int64("bytes_tx", stats.BytesTx),
			zap.Int("peers_count", stats.PeersCount))
	}
}

// checkPeersHealth проверяет здоровье пиров туннеля
func (m *MonitorService) checkPeersHealth(ctx context.Context, tunnel *domain.Tunnel, peersHealth []domain.PeerHealth) {
	for _, peerHealth := range peersHealth {
		// Получаем пира
		peer, err := m.peerManager.GetPeer(ctx, tunnel.ID, peerHealth.PeerID)
		if err != nil {
			m.logger.Error("failed to get peer for health check",
				zap.String("tunnel_id", tunnel.ID),
				zap.String("peer_id", peerHealth.PeerID),
				zap.Error(err))
			continue
		}

		// Обновляем статус пира
		peer.Status = peerHealth.Status
		peer.LastSeen = time.Now()
		peer.Latency = peerHealth.Latency
		peer.PacketLoss = peerHealth.PacketLoss
		peer.ConnectionQuality = peerHealth.ConnectionQuality
		peer.UpdatedAt = time.Now()

		// Если пир неактивен, логируем это
		if peer.Status == domain.PeerStatusOffline || peer.Status == domain.PeerStatusError {
			m.logger.Warn("peer health check failed",
				zap.String("tunnel_id", tunnel.ID),
				zap.String("peer_id", peer.ID),
				zap.String("status", string(peer.Status)),
				zap.Duration("latency", peer.Latency),
				zap.Float64("packet_loss", peer.PacketLoss))
		}
	}
}
