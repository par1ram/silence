package bypass

import (
	"testing"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMockBypassAdapter(t *testing.T) {
	logger := zap.NewNop()
	mock := NewMockBypassAdapter(logger)

	t.Run("создание mock адаптера", func(t *testing.T) {
		assert.NotNil(t, mock)
		assert.NotNil(t, mock.logger)
		assert.NotNil(t, mock.running)
		assert.NotNil(t, mock.stats)
		assert.Empty(t, mock.running)
		assert.Empty(t, mock.stats)
	})

	t.Run("запуск bypass соединения", func(t *testing.T) {
		config := &domain.BypassConfig{
			ID:         "test-1",
			Name:       "Test Bypass",
			Method:     domain.BypassMethodShadowsocks,
			LocalPort:  8080,
			RemoteHost: "example.com",
			RemotePort: 443,
			Password:   "test123",
			Encryption: "aes-256-gcm",
		}

		err := mock.Start(config)
		assert.NoError(t, err)
		assert.True(t, mock.IsRunning("test-1"))
		assert.Contains(t, mock.running, "test-1")
		assert.Contains(t, mock.stats, "test-1")
	})

	t.Run("запуск дублирующегося соединения", func(t *testing.T) {
		config := &domain.BypassConfig{
			ID:         "test-1", // тот же ID
			Name:       "Test Bypass 2",
			Method:     domain.BypassMethodV2Ray,
			LocalPort:  8081,
			RemoteHost: "example2.com",
			RemotePort: 443,
		}

		err := mock.Start(config)
		assert.NoError(t, err) // Mock позволяет дублирующиеся соединения
		assert.True(t, mock.IsRunning("test-1"))
	})

	t.Run("проверка IsRunning для несуществующего соединения", func(t *testing.T) {
		running := mock.IsRunning("nonexistent")
		assert.False(t, running)
	})

	t.Run("получение статистики", func(t *testing.T) {
		// Сначала запускаем соединение
		config := &domain.BypassConfig{
			ID:         "test-2",
			Name:       "Test Bypass 2",
			Method:     domain.BypassMethodObfs4,
			LocalPort:  8082,
			RemoteHost: "example3.com",
			RemotePort: 443,
		}
		err := mock.Start(config)
		assert.NoError(t, err)

		// Получаем статистику
		stats, err := mock.GetStats("test-2")
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, "test-2", stats.ID)
		assert.Greater(t, stats.BytesRx, int64(0))
		assert.Greater(t, stats.BytesTx, int64(0))
		assert.True(t, stats.LastActivity.After(time.Now().Add(-time.Second)))
	})

	t.Run("получение статистики несуществующего соединения", func(t *testing.T) {
		stats, err := mock.GetStats("nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, stats)
	})

	t.Run("остановка соединения", func(t *testing.T) {
		// Сначала запускаем соединение
		config := &domain.BypassConfig{
			ID:         "test-3",
			Name:       "Test Bypass 3",
			Method:     domain.BypassMethodCustom,
			LocalPort:  8083,
			RemoteHost: "example4.com",
			RemotePort: 443,
		}
		err := mock.Start(config)
		assert.NoError(t, err)
		assert.True(t, mock.IsRunning("test-3"))

		// Останавливаем соединение
		err = mock.Stop("test-3")
		assert.NoError(t, err)
		assert.False(t, mock.IsRunning("test-3"))
		assert.NotContains(t, mock.running, "test-3")
	})

	t.Run("остановка несуществующего соединения", func(t *testing.T) {
		err := mock.Stop("nonexistent")
		assert.NoError(t, err) // Mock не возвращает ошибку для несуществующих соединений
	})

	t.Run("множественные соединения", func(t *testing.T) {
		// Запускаем несколько соединений
		configs := []*domain.BypassConfig{
			{
				ID:         "multi-1",
				Name:       "Multi 1",
				Method:     domain.BypassMethodShadowsocks,
				LocalPort:  9001,
				RemoteHost: "multi1.com",
				RemotePort: 443,
			},
			{
				ID:         "multi-2",
				Name:       "Multi 2",
				Method:     domain.BypassMethodV2Ray,
				LocalPort:  9002,
				RemoteHost: "multi2.com",
				RemotePort: 443,
			},
		}

		for _, config := range configs {
			err := mock.Start(config)
			assert.NoError(t, err)
			assert.True(t, mock.IsRunning(config.ID))
		}

		// Проверяем, что все соединения запущены
		assert.True(t, mock.IsRunning("multi-1"))
		assert.True(t, mock.IsRunning("multi-2"))

		// Останавливаем одно соединение
		err := mock.Stop("multi-1")
		assert.NoError(t, err)
		assert.False(t, mock.IsRunning("multi-1"))
		assert.True(t, mock.IsRunning("multi-2"))
	})
}
