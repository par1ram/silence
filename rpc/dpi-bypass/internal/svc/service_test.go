package svc

import (
	"testing"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/bypass"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/adapters/grpc"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/config"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServiceContext(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "info",
		Version:  "1.0.0",
		GRPC: config.GRPCConfig{
			Address: ":9091",
		},
	}
	logger := zap.NewNop()

	svcCtx := NewServiceContext(cfg, logger)

	assert.NotNil(t, svcCtx)
	assert.Equal(t, cfg, svcCtx.Config)
	assert.Equal(t, logger, svcCtx.Logger)
	assert.IsType(t, &services.HealthService{}, svcCtx.HealthService)
	assert.IsType(t, &services.BypassService{}, svcCtx.BypassService)
	assert.IsType(t, &bypass.MultiBypassAdapter{}, svcCtx.BypassAdapter)
	assert.IsType(t, &grpc.Server{}, svcCtx.GRPCServer)
}
