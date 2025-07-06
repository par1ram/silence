package bypass

import (
	"net"
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestV2RayAdapter_Basic(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewV2RayAdapter(logger)
	assert.NotNil(t, adapter)

	config := &domain.BypassConfig{
		ID:         "v2-1",
		Name:       "V2Ray Test",
		Method:     domain.BypassMethodV2Ray,
		LocalPort:  0,
		RemoteHost: "localhost",
		RemotePort: 12345,
		Password:   "testpass",
		Encryption: "none",
	}

	err := adapter.Start(config)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning("v2-1"))

	stats, err := adapter.GetStats("v2-1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "v2-1", stats.ID)

	err = adapter.Stop("v2-1")
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning("v2-1"))

	stats, err = adapter.GetStats("v2-1")
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestV2RayAdapter_Start_PortBusy(t *testing.T) {
	logger := zap.NewNop()
	adapter1 := NewV2RayAdapter(logger)
	adapter2 := NewV2RayAdapter(logger)

	config := &domain.BypassConfig{
		ID:         "v2-busy-1",
		Name:       "Busy Test",
		Method:     domain.BypassMethodV2Ray,
		LocalPort:  0,
		RemoteHost: "localhost",
		RemotePort: 12345,
		Password:   "testpass",
		Encryption: "none",
	}

	err := adapter1.Start(config)
	assert.NoError(t, err)

	adapter1.mutex.RLock()
	conn := adapter1.running["v2-busy-1"]
	adapter1.mutex.RUnlock()
	port := conn.listener.Addr().(*net.TCPAddr).Port

	config2 := *config
	config2.ID = "v2-busy-2"
	config2.LocalPort = port
	err = adapter2.Start(&config2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create TCP listener")

	err = adapter1.Stop("v2-busy-1")
	assert.NoError(t, err)
}
