package services

import (
	"testing"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPeerService_calculateConnectionQuality(t *testing.T) {
	logger := zap.NewNop()
	ps := NewPeerService(logger).(*PeerService)

	t.Run("качество соединения для активного пира", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer1",
			Status:        domain.PeerStatusActive,
			LastHandshake: time.Now().Add(-1 * time.Minute), // недавний handshake
			Latency:       50 * time.Millisecond,
			PacketLoss:    0.01,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Greater(t, quality, 0.9) // высокое качество
	})

	t.Run("качество соединения для неактивного пира", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer2",
			Status:        domain.PeerStatusInactive,
			LastHandshake: time.Now().Add(-3 * time.Minute),
			Latency:       100 * time.Millisecond,
			PacketLoss:    0.05,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.5) // низкое качество из-за статуса
	})

	t.Run("качество соединения для оффлайн пира", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer3",
			Status:        domain.PeerStatusOffline,
			LastHandshake: time.Now().Add(-15 * time.Minute), // очень старый handshake
			Latency:       1000 * time.Millisecond,
			PacketLoss:    0.2,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.2) // очень низкое качество
	})

	t.Run("качество соединения с высоким packet loss", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer4",
			Status:        domain.PeerStatusActive,
			LastHandshake: time.Now().Add(-30 * time.Second),
			Latency:       50 * time.Millisecond,
			PacketLoss:    0.5, // 50% потерь
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.6) // качество снижено из-за потерь
	})

	t.Run("качество соединения с высокой задержкой", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer5",
			Status:        domain.PeerStatusActive,
			LastHandshake: time.Now().Add(-1 * time.Minute),
			Latency:       600 * time.Millisecond, // высокая задержка
			PacketLoss:    0.01,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.6) // качество снижено из-за задержки
	})

	t.Run("качество соединения без handshake", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer6",
			Status:        domain.PeerStatusActive,
			LastHandshake: time.Time{}, // нулевое время
			Latency:       50 * time.Millisecond,
			PacketLoss:    0.0, // убираем packet loss для точного теста
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Equal(t, 1.0, quality) // максимальное качество без учета handshake
	})

	t.Run("качество соединения с очень старым handshake", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer7",
			Status:        domain.PeerStatusActive,
			LastHandshake: time.Now().Add(-15 * time.Minute), // очень старый
			Latency:       50 * time.Millisecond,
			PacketLoss:    0.01,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.2) // очень низкое качество из-за старого handshake
	})

	t.Run("качество соединения с пиром в ошибке", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer8",
			Status:        domain.PeerStatusError,
			LastHandshake: time.Now().Add(-1 * time.Minute),
			Latency:       50 * time.Millisecond,
			PacketLoss:    0.01,
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.2) // очень низкое качество из-за статуса ошибки
	})

	t.Run("качество соединения с максимальными проблемами", func(t *testing.T) {
		peer := &domain.Peer{
			ID:            "peer9",
			Status:        domain.PeerStatusOffline,
			LastHandshake: time.Now().Add(-20 * time.Minute), // очень старый
			Latency:       1000 * time.Millisecond,           // высокая задержка
			PacketLoss:    0.8,                               // высокие потери
		}

		quality := ps.calculateConnectionQuality(peer)
		assert.Less(t, quality, 0.1) // минимальное качество
	})
}
