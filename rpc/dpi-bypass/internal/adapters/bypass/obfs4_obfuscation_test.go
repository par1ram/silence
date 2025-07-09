package bypass

import (
	"testing"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestObfs4Adapter_generateKey(t *testing.T) {
	o := &Obfs4Adapter{}
	k1 := o.generateKey("pass1")
	k2 := o.generateKey("pass1")
	k3 := o.generateKey("pass2")
	assert.Equal(t, k1, k2)
	assert.NotEqual(t, k1, k3)
	assert.Len(t, k1, 32)
}

func TestObfs4Adapter_obfuscateData(t *testing.T) {
	o := &Obfs4Adapter{}
	conn := &obfs4Connection{config: &domain.BypassConfig{Parameters: map[string]string{"password": "test"}}}
	plain := []byte("hello world")
	obf := o.obfuscateData(plain, conn)
	assert.NotEqual(t, plain, obf)
	// Повторное применение вернет исходные данные
	deobf := o.obfuscateData(obf, conn)
	assert.Equal(t, plain, deobf)
}

func TestObfs4Adapter_applyIATDelay(t *testing.T) {
	o := &Obfs4Adapter{}
	conn := &obfs4Connection{iatDist: "pareto", iatDistMin: 1, iatDistMax: 2}
	start := time.Now()
	o.applyIATDelay(conn)
	assert.GreaterOrEqual(t, time.Since(start), 0*time.Millisecond)

	conn.iatDist = "uniform"
	o.applyIATDelay(conn)
	assert.GreaterOrEqual(t, time.Since(start), 0*time.Millisecond)

	conn.iatDist = "unknown"
	o.applyIATDelay(conn)
	assert.GreaterOrEqual(t, time.Since(start), 0*time.Millisecond)
}

func TestObfs4Adapter_paretoDistribution(t *testing.T) {
	o := &Obfs4Adapter{}
	val := o.paretoDistribution(1, 10)
	assert.GreaterOrEqual(t, val, 1)
	assert.LessOrEqual(t, val, 10)
}

func TestObfs4Adapter_uniformDistribution(t *testing.T) {
	o := &Obfs4Adapter{}
	val := o.uniformDistribution(1, 10)
	assert.GreaterOrEqual(t, val, 1)
	assert.LessOrEqual(t, val, 10)
}

func TestObfs4Adapter_randomFloat(t *testing.T) {
	o := &Obfs4Adapter{}
	val := o.randomFloat()
	assert.GreaterOrEqual(t, val, 0.0)
	assert.LessOrEqual(t, val, 1.0)
}

func TestObfs4Adapter_statsHelpers(t *testing.T) {
	o := &Obfs4Adapter{}
	conn := &obfs4Connection{stats: &domain.BypassStats{}}
	o.updateStats(conn, 10, 20)
	assert.Equal(t, int64(10), conn.stats.BytesReceived)
	assert.Equal(t, int64(20), conn.stats.BytesSent)
	o.updateLastActivity(conn)
	assert.WithinDuration(t, time.Now(), conn.stats.EndTime, time.Second)
	o.incrementConnections(conn)
	assert.Equal(t, int64(1), conn.stats.ConnectionsEstablished)
	o.incrementErrorCount(conn)
	assert.Equal(t, int64(1), conn.stats.ConnectionsFailed)
}
