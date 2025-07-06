package services

import (
	"context"
	"errors"
	"testing"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTunnelService_DeleteStartStop(t *testing.T) {
	logger := zap.NewNop()
	keyGen := NewKeyGenerator()
	wgMock := &mockWGManager{}
	ts := NewTunnelService(keyGen, wgMock, logger).(*TunnelService)
	ctx := context.Background()

	// Создаем туннель
	tun, err := ts.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test", ListenPort: 51820, MTU: 1420})
	assert.NoError(t, err)
	assert.NotNil(t, tun)

	// Удаление несуществующего туннеля
	err = ts.DeleteTunnel(ctx, "not-exist")
	assert.Error(t, err)

	// Удаление существующего
	err = ts.DeleteTunnel(ctx, tun.ID)
	assert.NoError(t, err)

	// Повторное удаление
	err = ts.DeleteTunnel(ctx, tun.ID)
	assert.Error(t, err)
}

func TestTunnelService_StartStopTunnel(t *testing.T) {
	logger := zap.NewNop()
	keyGen := NewKeyGenerator()
	wgMock := &mockWGManager{}
	ts := NewTunnelService(keyGen, wgMock, logger).(*TunnelService)
	ctx := context.Background()

	tun, _ := ts.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test", ListenPort: 51820, MTU: 1420})

	// Запуск несуществующего туннеля
	err := ts.StartTunnel(ctx, "not-exist")
	assert.Error(t, err)

	// Ошибка создания интерфейса
	wgMock.CreateErr = errors.New("fail create")
	err = ts.StartTunnel(ctx, tun.ID)
	assert.Error(t, err)
	wgMock.CreateErr = nil

	// Успешный запуск
	err = ts.StartTunnel(ctx, tun.ID)
	assert.NoError(t, err)

	// Остановка несуществующего туннеля
	err = ts.StopTunnel(ctx, "not-exist")
	assert.Error(t, err)

	// Ошибка удаления интерфейса
	wgMock.DeleteErr = errors.New("fail delete")
	err = ts.StopTunnel(ctx, tun.ID)
	assert.Error(t, err)
	wgMock.DeleteErr = nil

	// Успешная остановка
	err = ts.StopTunnel(ctx, tun.ID)
	assert.NoError(t, err)
}

func TestTunnelMonitor_AutoRecoveryAndStats(t *testing.T) {
	logger := zap.NewNop()
	keyGen := NewKeyGenerator()
	wgMock := &mockWGManager{}
	ts := NewTunnelService(keyGen, wgMock, logger).(*TunnelService)
	ctx := context.Background()

	tun, _ := ts.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test", ListenPort: 51820, MTU: 1420})

	// EnableAutoRecovery для несуществующего
	err := ts.EnableAutoRecovery(ctx, "not-exist")
	assert.Error(t, err)

	// EnableAutoRecovery для существующего
	err = ts.EnableAutoRecovery(ctx, tun.ID)
	assert.NoError(t, err)
	assert.True(t, ts.tunnels[tun.ID].AutoRecovery)

	// DisableAutoRecovery для несуществующего
	err = ts.DisableAutoRecovery(ctx, "not-exist")
	assert.Error(t, err)

	// DisableAutoRecovery для существующего
	err = ts.DisableAutoRecovery(ctx, tun.ID)
	assert.NoError(t, err)
	assert.False(t, ts.tunnels[tun.ID].AutoRecovery)

	// GetTunnelStats для несуществующего
	_, err = ts.GetTunnelStats(ctx, "not-exist")
	assert.Error(t, err)

	// GetTunnelStats для существующего
	wgMock.Stats = &ports.InterfaceStats{BytesRx: 123, BytesTx: 456}
	stats, err := ts.GetTunnelStats(ctx, tun.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), stats.BytesRx)
	assert.Equal(t, int64(456), stats.BytesTx)

	// HealthCheck для несуществующего
	_, err = ts.HealthCheck(ctx, &domain.HealthCheckRequest{TunnelID: "not-exist"})
	assert.Error(t, err)

	// HealthCheck для существующего
	resp, err := ts.HealthCheck(ctx, &domain.HealthCheckRequest{TunnelID: tun.ID})
	assert.NoError(t, err)
	assert.Equal(t, tun.ID, resp.TunnelID)

	// RecoverTunnel для несуществующего
	err = ts.RecoverTunnel(ctx, "not-exist")
	assert.Error(t, err)

	// RecoverTunnel для существующего (ошибка создания интерфейса)
	wgMock.CreateErr = errors.New("fail create")
	err = ts.RecoverTunnel(ctx, tun.ID)
	assert.Error(t, err)
	wgMock.CreateErr = nil

	// RecoverTunnel для существующего (успех)
	err = ts.RecoverTunnel(ctx, tun.ID)
	assert.NoError(t, err)
}
