package bypass

import (
	"fmt"
	"net"
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCustomAdapter_Basic(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewCustomAdapter(logger)
	assert.NotNil(t, adapter)

	config := &domain.BypassConfig{
		ID:     "custom-1",
		Name:   "Custom Test",
		Method: domain.BypassMethodCustom,
		Parameters: map[string]string{
			"local_port":  "0", // 0 — выбрать свободный порт
			"remote_host": "localhost",
			"remote_port": "12345",
			"password":    "testpass",
			"encryption":  "none",
		},
	}

	err := adapter.Start(config)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning("custom-1"))

	stats, err := adapter.GetStats("custom-1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "custom-1", stats.ID)

	err = adapter.Stop("custom-1")
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning("custom-1"))

	stats, err = adapter.GetStats("custom-1")
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestCustomAdapter_Start_PortBusy(t *testing.T) {
	logger := zap.NewNop()
	adapter1 := NewCustomAdapter(logger)
	adapter2 := NewCustomAdapter(logger)

	config := &domain.BypassConfig{
		ID:     "busy-1",
		Name:   "Busy Test",
		Method: domain.BypassMethodCustom,
		Parameters: map[string]string{
			"local_port":  "0", // 0 — выбрать свободный порт
			"remote_host": "localhost",
			"remote_port": "12345",
			"password":    "testpass",
			"encryption":  "none",
		},
	}

	err := adapter1.Start(config)
	assert.NoError(t, err)

	// Получаем реальный порт, который был выбран
	adapter1.mutex.RLock()
	conn := adapter1.running["busy-1"]
	adapter1.mutex.RUnlock()

	if conn == nil || conn.listener == nil {
		t.Fatal("connection or listener is nil")
	}

	port := conn.listener.Addr().(*net.TCPAddr).Port

	// Пытаемся запустить второй адаптер на том же порту
	config2 := *config
	config2.ID = "busy-2"
	config2.Parameters["local_port"] = fmt.Sprintf("%d", port)
	err = adapter2.Start(&config2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create listener")

	err = adapter1.Stop("busy-1")
	assert.NoError(t, err)
}
