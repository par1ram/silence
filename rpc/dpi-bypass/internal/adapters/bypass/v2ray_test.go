package bypass

import (
	"fmt"
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
		ID:     "v2ray-1",
		Name:   "V2Ray Test",
		Method: domain.BypassMethodV2Ray,
		Parameters: map[string]string{
			"local_port":  "0",
			"remote_host": "localhost",
			"remote_port": "12345",
			"password":    "testpass",
			"encryption":  "none",
		},
	}

	err := adapter.Start(config)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning("v2ray-1"))

	stats, err := adapter.GetStats("v2ray-1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "v2ray-1", stats.ID)

	err = adapter.Stop("v2ray-1")
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning("v2ray-1"))

	stats, err = adapter.GetStats("v2ray-1")
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestV2RayAdapter_Start_PortBusy(t *testing.T) {
	logger := zap.NewNop()
	adapter1 := NewV2RayAdapter(logger)
	adapter2 := NewV2RayAdapter(logger)

	config := &domain.BypassConfig{
		ID:     "v2ray-busy-1",
		Name:   "Busy Test",
		Method: domain.BypassMethodV2Ray,
		Parameters: map[string]string{
			"local_port":  "0",
			"remote_host": "localhost",
			"remote_port": "12345",
			"password":    "testpass",
			"encryption":  "none",
		},
	}

	err := adapter1.Start(config)
	assert.NoError(t, err)

	adapter1.mutex.RLock()
	conn := adapter1.running["v2ray-busy-1"]
	adapter1.mutex.RUnlock()

	if conn == nil || conn.listener == nil {
		t.Fatal("connection or listener is nil")
	}

	port := conn.listener.Addr().(*net.TCPAddr).Port

	config2 := *config
	config2.ID = "v2ray-busy-2"
	config2.Parameters["local_port"] = fmt.Sprintf("%d", port)
	err = adapter2.Start(&config2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create listener")

	err = adapter1.Stop("v2ray-busy-1")
	assert.NoError(t, err)
}
