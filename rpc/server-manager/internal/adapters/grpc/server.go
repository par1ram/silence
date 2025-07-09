package grpc

import (
	"context"
	"net"

	"github.com/par1ram/silence/rpc/server-manager/api/proto"
	"github.com/par1ram/silence/rpc/server-manager/internal/config"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server gRPC сервер для server-manager сервиса
type Server struct {
	server        *grpc.Server
	listener      net.Listener
	serverService ports.ServerService
	logger        *zap.Logger
	config        *config.Config
}

// NewServer создает новый gRPC сервер
func NewServer(
	serverService ports.ServerService,
	logger *zap.Logger,
	cfg *config.Config,
) *Server {
	return &Server{
		serverService: serverService,
		logger:        logger,
		config:        cfg,
	}
}

// Start запускает gRPC сервер
func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.config.GRPC.Address)
	if err != nil {
		return err
	}

	s.listener = listener

	// Создаем gRPC сервер
	s.server = grpc.NewServer()

	// Регистрируем сервис
	serverHandler := NewServerManagerHandler(s.serverService, s.logger)
	proto.RegisterServerManagerServiceServer(s.server, serverHandler)

	// Включаем reflection для отладки
	reflection.Register(s.server)

	s.logger.Info("gRPC server starting", zap.String("address", s.config.GRPC.Address))

	go func() {
		<-ctx.Done()
		s.logger.Info("gRPC server shutting down")
		s.server.GracefulStop()
	}()

	return s.server.Serve(listener)
}

// Stop останавливает gRPC сервер
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Name возвращает имя сервиса
func (s *Server) Name() string {
	return "grpc-server"
}
