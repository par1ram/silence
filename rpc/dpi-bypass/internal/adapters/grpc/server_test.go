package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/par1ram/silence/rpc/dpi-bypass/api/proto"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/config"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func TestServer_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Address: ":0", // Use any available port
		},
	}

	server := NewServer(mockService, logger, cfg)

	// Test server name
	assert.Equal(t, "grpc-server", server.Name())

	// Test server stop without starting
	err := server.Stop(context.Background())
	assert.NoError(t, err)
}

func TestServer_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()

	// Create a buffer connection for testing
	lis := bufconn.Listen(bufSize)

	// Create gRPC server
	s := grpc.NewServer()

	// Register service
	dpiHandler := NewDPIBypassHandler(mockService, logger)
	proto.RegisterDpiBypassServiceServer(s, dpiHandler)

	// Start server in background
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Create client connection
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	// Create client
	client := proto.NewDpiBypassServiceClient(conn)

	// Test Health endpoint
	t.Run("Health", func(t *testing.T) {
		req := &proto.HealthRequest{}
		resp, err := client.Health(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "healthy", resp.Status)
		assert.Equal(t, "1.0.0", resp.Version)
	})

	// Test CreateBypassConfig endpoint
	t.Run("CreateBypassConfig", func(t *testing.T) {
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

		// Mock the service call
		mockService.EXPECT().
			CreateBypassConfig(gomock.Any(), gomock.Any()).
			Return(&domain.BypassConfig{
				ID:          "test-id",
				Name:        req.Name,
				Description: req.Description,
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodProxyChain,
				Status:      domain.BypassStatusInactive,
				Parameters:  req.Parameters,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil)

		resp, err := client.CreateBypassConfig(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-id", resp.Id)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.Description, resp.Description)
	})

	// Cleanup
	s.GracefulStop()
}

func TestServer_StartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Address: "invalid-address", // Invalid address to trigger error
		},
	}

	server := NewServer(mockService, logger, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	assert.Error(t, err)
}

func TestServer_StopWithoutStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Address: ":0",
		},
	}

	server := NewServer(mockService, logger, cfg)

	// Stop server without starting it
	err := server.Stop(context.Background())
	assert.NoError(t, err)
}

func TestServer_Name(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockDPIBypassService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Address: ":0",
		},
	}

	server := NewServer(mockService, logger, cfg)
	assert.Equal(t, "grpc-server", server.Name())
}
