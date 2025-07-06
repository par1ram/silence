package bypass

import (
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAdapterFactory(t *testing.T) {
	logger := zap.NewNop()
	factory := NewAdapterFactory(logger)

	t.Run("создание фабрики", func(t *testing.T) {
		assert.NotNil(t, factory)
		assert.NotNil(t, factory.logger)
	})

	t.Run("создание Shadowsocks адаптера", func(t *testing.T) {
		adapter, err := factory.CreateAdapter(domain.BypassMethodShadowsocks)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.IsType(t, &ShadowsocksAdapter{}, adapter)
	})

	t.Run("создание V2Ray адаптера", func(t *testing.T) {
		adapter, err := factory.CreateAdapter(domain.BypassMethodV2Ray)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.IsType(t, &V2RayAdapter{}, adapter)
	})

	t.Run("создание Obfs4 адаптера", func(t *testing.T) {
		adapter, err := factory.CreateAdapter(domain.BypassMethodObfs4)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.IsType(t, &Obfs4Adapter{}, adapter)
	})

	t.Run("создание Custom адаптера", func(t *testing.T) {
		adapter, err := factory.CreateAdapter(domain.BypassMethodCustom)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.IsType(t, &CustomAdapter{}, adapter)
	})

	t.Run("неподдерживаемый метод", func(t *testing.T) {
		adapter, err := factory.CreateAdapter("unsupported")
		assert.Error(t, err)
		assert.Nil(t, adapter)
		assert.Contains(t, err.Error(), "unsupported bypass method")
	})
}

func TestMultiBypassAdapter(t *testing.T) {
	logger := zap.NewNop()
	multiAdapter := NewMultiBypassAdapter(logger)

	t.Run("создание мульти-адаптера", func(t *testing.T) {
		assert.NotNil(t, multiAdapter)
		assert.NotNil(t, multiAdapter.logger)
		assert.NotNil(t, multiAdapter.adapters)
		assert.Empty(t, multiAdapter.adapters)
	})

	t.Run("проверка IsRunning для несуществующего соединения", func(t *testing.T) {
		running := multiAdapter.IsRunning("nonexistent")
		assert.False(t, running)
	})

	t.Run("получение статистики несуществующего соединения", func(t *testing.T) {
		stats, err := multiAdapter.GetStats("nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, stats)
	})

	t.Run("остановка несуществующего соединения", func(t *testing.T) {
		err := multiAdapter.Stop("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bypass connection not found")
	})
}
