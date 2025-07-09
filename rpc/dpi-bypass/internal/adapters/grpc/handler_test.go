package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/par1ram/silence/rpc/dpi-bypass/api/proto"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDPIBypassHandler_Health(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.HealthRequest{}
	resp, err := handler.Health(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.NotNil(t, resp.Timestamp)
}

func TestDPIBypassHandler_CreateBypassConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.CreateBypassConfigRequest{
		Name:        "test-config",
		Description: "test configuration",
		Type:        proto.BypassType_BYPASS_TYPE_TUNNEL_OBFUSCATION,
		Method:      proto.BypassMethod_BYPASS_METHOD_PROXY_CHAIN,
		Parameters: map[string]string{
			"local_port":  "1080",
			"remote_host": "example.com",
			"remote_port": "8080",
		},
	}

	expectedConfig := &domain.BypassConfig{
		ID:          "test-id",
		Name:        req.Name,
		Description: req.Description,
		Type:        domain.BypassTypeTunnelObfuscation,
		Method:      domain.BypassMethodProxyChain,
		Status:      domain.BypassStatusInactive,
		Parameters:  req.Parameters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.EXPECT().
		CreateBypassConfig(gomock.Any(), gomock.Any()).
		Return(expectedConfig, nil)

	resp, err := handler.CreateBypassConfig(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedConfig.ID, resp.Id)
	assert.Equal(t, expectedConfig.Name, resp.Name)
	assert.Equal(t, expectedConfig.Description, resp.Description)
}

func TestDPIBypassHandler_CreateBypassConfig_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.CreateBypassConfigRequest{
		Name:        "test-config",
		Description: "test configuration",
		Type:        proto.BypassType_BYPASS_TYPE_TUNNEL_OBFUSCATION,
		Method:      proto.BypassMethod_BYPASS_METHOD_PROXY_CHAIN,
	}

	mockService.EXPECT().
		CreateBypassConfig(gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	resp, err := handler.CreateBypassConfig(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestDPIBypassHandler_GetBypassConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	configID := "test-id"
	req := &proto.GetBypassConfigRequest{
		Id: configID,
	}

	expectedConfig := &domain.BypassConfig{
		ID:          configID,
		Name:        "test-config",
		Description: "test configuration",
		Type:        domain.BypassTypeTunnelObfuscation,
		Method:      domain.BypassMethodProxyChain,
		Status:      domain.BypassStatusInactive,
		Parameters:  map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.EXPECT().
		GetBypassConfig(gomock.Any(), configID).
		Return(expectedConfig, nil)

	resp, err := handler.GetBypassConfig(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedConfig.ID, resp.Id)
	assert.Equal(t, expectedConfig.Name, resp.Name)
}

func TestDPIBypassHandler_GetBypassConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.GetBypassConfigRequest{
		Id: "non-existent-id",
	}

	mockService.EXPECT().
		GetBypassConfig(gomock.Any(), "non-existent-id").
		Return(nil, assert.AnError)

	resp, err := handler.GetBypassConfig(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestDPIBypassHandler_ListBypassConfigs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.ListBypassConfigsRequest{
		Type:   proto.BypassType_BYPASS_TYPE_TUNNEL_OBFUSCATION,
		Status: proto.BypassStatus_BYPASS_STATUS_INACTIVE,
		Limit:  10,
		Offset: 0,
	}

	expectedConfigs := []*domain.BypassConfig{
		{
			ID:          "config-1",
			Name:        "config-1",
			Description: "first config",
			Type:        domain.BypassTypeTunnelObfuscation,
			Method:      domain.BypassMethodProxyChain,
			Status:      domain.BypassStatusInactive,
			Parameters:  map[string]string{"key": "value"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "config-2",
			Name:        "config-2",
			Description: "second config",
			Type:        domain.BypassTypeTunnelObfuscation,
			Method:      domain.BypassMethodProxyChain,
			Status:      domain.BypassStatusInactive,
			Parameters:  map[string]string{"key": "value"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockService.EXPECT().
		ListBypassConfigs(gomock.Any(), gomock.Any()).
		Return(expectedConfigs, 2, nil)

	resp, err := handler.ListBypassConfigs(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Configs, 2)
	assert.Equal(t, int32(2), resp.Total)
	assert.Equal(t, expectedConfigs[0].ID, resp.Configs[0].Id)
	assert.Equal(t, expectedConfigs[1].ID, resp.Configs[1].Id)
}

func TestDPIBypassHandler_StartBypass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.StartBypassRequest{
		ConfigId:   "test-config-id",
		TargetHost: "example.com",
		TargetPort: 80,
		Options: map[string]string{
			"timeout": "30s",
		},
	}

	expectedSession := &domain.BypassSession{
		ID:         "session-123",
		ConfigID:   req.ConfigId,
		TargetHost: req.TargetHost,
		TargetPort: int(req.TargetPort),
		Status:     domain.BypassStatusActive,
		StartedAt:  time.Now(),
	}

	mockService.EXPECT().
		StartBypass(gomock.Any(), gomock.Any()).
		Return(expectedSession, nil)

	resp, err := handler.StartBypass(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, expectedSession.ID, resp.SessionId)
	assert.Contains(t, resp.Message, "started successfully")
}

func TestDPIBypassHandler_StartBypass_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.StartBypassRequest{
		ConfigId:   "test-config-id",
		TargetHost: "example.com",
		TargetPort: 80,
	}

	mockService.EXPECT().
		StartBypass(gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	resp, err := handler.StartBypass(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestDPIBypassHandler_StopBypass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.StopBypassRequest{
		SessionId: "session-123",
	}

	mockService.EXPECT().
		StopBypass(gomock.Any(), req.SessionId).
		Return(nil)

	resp, err := handler.StopBypass(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "stopped successfully")
}

func TestDPIBypassHandler_StopBypass_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.StopBypassRequest{
		SessionId: "session-123",
	}

	mockService.EXPECT().
		StopBypass(gomock.Any(), req.SessionId).
		Return(assert.AnError)

	resp, err := handler.StopBypass(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestDPIBypassHandler_GetBypassStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.GetBypassStatsRequest{
		SessionId: "session-123",
	}

	expectedStats := &domain.BypassStats{
		ID:                     "stats-123",
		ConfigID:               "config-123",
		SessionID:              "session-123",
		BytesSent:              1024,
		BytesReceived:          2048,
		PacketsSent:            10,
		PacketsReceived:        20,
		ConnectionsEstablished: 1,
		ConnectionsFailed:      0,
		SuccessRate:            100.0,
		AverageLatency:         50.0,
		StartTime:              time.Now(),
		EndTime:                time.Now(),
	}

	mockService.EXPECT().
		GetBypassStats(gomock.Any(), req.SessionId).
		Return(expectedStats, nil)

	resp, err := handler.GetBypassStats(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedStats.ID, resp.Id)
	assert.Equal(t, expectedStats.SessionID, resp.SessionId)
	assert.Equal(t, expectedStats.BytesSent, resp.BytesSent)
	assert.Equal(t, expectedStats.BytesReceived, resp.BytesReceived)
	assert.Equal(t, expectedStats.SuccessRate, resp.SuccessRate)
}

func TestDPIBypassHandler_DeleteBypassConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.DeleteBypassConfigRequest{
		Id: "config-123",
	}

	mockService.EXPECT().
		DeleteBypassConfig(gomock.Any(), req.Id).
		Return(nil)

	resp, err := handler.DeleteBypassConfig(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestDPIBypassHandler_DeleteBypassConfig_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	handler := NewDPIBypassHandler(mockService, logger)

	req := &proto.DeleteBypassConfigRequest{
		Id: "config-123",
	}

	mockService.EXPECT().
		DeleteBypassConfig(gomock.Any(), req.Id).
		Return(assert.AnError)

	resp, err := handler.DeleteBypassConfig(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}
