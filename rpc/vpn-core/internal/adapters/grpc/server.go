package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server gRPC сервер
type Server struct {
	server        *grpc.Server
	port          string
	logger        *zap.Logger
	tunnelManager ports.TunnelManager
	peerManager   ports.PeerManager
}

// NewServer создает новый gRPC сервер
func NewServer(port string, tunnelManager ports.TunnelManager, peerManager ports.PeerManager, logger *zap.Logger) *Server {
	server := grpc.NewServer()

	// Регистрируем сервис
	proto.RegisterVpnCoreServiceServer(server, &VpnCoreService{
		tunnelManager: tunnelManager,
		peerManager:   peerManager,
		logger:        logger,
	})

	// Включаем reflection для grpcurl
	reflection.Register(server)

	return &Server{
		server:        server,
		port:          port,
		logger:        logger,
		tunnelManager: tunnelManager,
		peerManager:   peerManager,
	}
}

// Start запускает gRPC сервер
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("starting gRPC server", zap.String("port", s.port))
	return s.server.Serve(lis)
}

// Stop останавливает gRPC сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping gRPC server")
	s.server.GracefulStop()
	return nil
}

// Name возвращает имя сервиса
func (s *Server) Name() string {
	return "vpn-core-grpc"
}
