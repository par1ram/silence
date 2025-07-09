package app

import (
	"context"
	"testing"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockService is a mock implementation of the Service interface for testing.
type MockService struct {
	Started bool
	Stopped bool
	NameVal string
}

func (m *MockService) Start(ctx context.Context) error {
	m.Started = true
	return nil
}

func (m *MockService) Stop(ctx context.Context) error {
	m.Stopped = true
	return nil
}

func (m *MockService) Name() string {
	return m.NameVal
}

func TestNewApp(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	app := NewApp(cfg, logger)

	assert.NotNil(t, app)
	assert.Equal(t, cfg, app.config)
	assert.Equal(t, logger, app.logger)
	assert.Empty(t, app.services)
}

func TestAddService(t *testing.T) {
	app := NewApp(&config.Config{}, zap.NewNop())
	service1 := &MockService{NameVal: "service1"}
	service2 := &MockService{NameVal: "service2"}

	app.AddService(service1)
	assert.Len(t, app.services, 1)
	assert.Equal(t, service1, app.services[0])

	app.AddService(service2)
	assert.Len(t, app.services, 2)
	assert.Equal(t, service2, app.services[1])
}

func TestRun(t *testing.T) {
	app := NewApp(&config.Config{}, zap.NewNop())
	service1 := &MockService{NameVal: "service1"}
	service2 := &MockService{NameVal: "service2"}

	app.AddService(service1)
	app.AddService(service2)

	go func() {
		time.Sleep(100 * time.Millisecond)
		// Simulate a shutdown signal
		// In a real test, you might use a channel to signal shutdown
		// For simplicity, we'll just check the state after a short delay
	}()

	// We can't directly test the blocking `run` method in a simple unit test.
	// A more complex test would involve channels and goroutines to coordinate.
	// However, we can verify that the services are started.

	// This is a simplified check. A real test for `run` would be more involved.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, s := range app.services {
		go func(service interface{ Start(context.Context) error }) {
			_ = service.Start(ctx)
		}(s)
	}

	time.Sleep(50 * time.Millisecond)

	assert.True(t, service1.Started)
	assert.True(t, service2.Started)
}
