package bypass

import (
	"fmt"
	"net"
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestObfs4Adapter_Basic(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewObfs4Adapter(logger)
	assert.NotNil(t, adapter)

	config := &domain.BypassConfig{
		ID:     "obfs4-1",
		Name:   "Obfs4 Test",
		Method: domain.BypassMethodObfs4,
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
	assert.True(t, adapter.IsRunning("obfs4-1"))

	stats, err := adapter.GetStats("obfs4-1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "obfs4-1", stats.ID)

	err = adapter.Stop("obfs4-1")
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning("obfs4-1"))

	stats, err = adapter.GetStats("obfs4-1")
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestObfs4Adapter_Start_PortBusy(t *testing.T) {
	logger := zap.NewNop()
	adapter1 := NewObfs4Adapter(logger)
	adapter2 := NewObfs4Adapter(logger)

	config := &domain.BypassConfig{
		ID:     "obfs4-busy-1",
		Name:   "Busy Test",
		Method: domain.BypassMethodObfs4,
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
	conn := adapter1.running["obfs4-busy-1"]
	adapter1.mutex.RUnlock()
	port := conn.listener.Addr().(*net.TCPAddr).Port

	config2 := *config
	config2.ID = "obfs4-busy-2"
	config2.Parameters["local_port"] = fmt.Sprintf("%d", port)
	err = adapter2.Start(&config2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create listener")

	err = adapter1.Stop("obfs4-busy-1")
	assert.NoError(t, err)
}
