package grpc

import (
	"context"

	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Health проверка здоровья сервиса
func (s *VpnCoreService) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	s.logger.Debug("health check requested")

	return &proto.HealthResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}
