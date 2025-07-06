package bypass

import (
	"net"
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestShadowsocksAdapter_Basic(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewShadowsocksAdapter(logger)
	assert.NotNil(t, adapter)

	config := &domain.BypassConfig{
		ID:         "ss-1",
		Name:       "Shadowsocks Test",
		Method:     domain.BypassMethodShadowsocks,
		LocalPort:  0,
		RemoteHost: "localhost",
		RemotePort: 12345,
		Password:   "testpass",
		Encryption: "none",
	}

	err := adapter.Start(config)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning("ss-1"))

	stats, err := adapter.GetStats("ss-1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "ss-1", stats.ID)

	err = adapter.Stop("ss-1")
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning("ss-1"))

	stats, err = adapter.GetStats("ss-1")
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestShadowsocksAdapter_Start_PortBusy(t *testing.T) {
	logger := zap.NewNop()
	adapter1 := NewShadowsocksAdapter(logger)
	adapter2 := NewShadowsocksAdapter(logger)

	config := &domain.BypassConfig{
		ID:         "ss-busy-1",
		Name:       "Busy Test",
		Method:     domain.BypassMethodShadowsocks,
		LocalPort:  0,
		RemoteHost: "localhost",
		RemotePort: 12345,
		Password:   "testpass",
		Encryption: "none",
	}

	err := adapter1.Start(config)
	assert.NoError(t, err)

	adapter1.mutex.RLock()
	conn := adapter1.running["ss-busy-1"]
	adapter1.mutex.RUnlock()
	port := conn.listener.Addr().(*net.TCPAddr).Port

	config2 := *config
	config2.ID = "ss-busy-2"
	config2.LocalPort = port
	err = adapter2.Start(&config2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create listener")

	err = adapter1.Stop("ss-busy-1")
	assert.NoError(t, err)
}
