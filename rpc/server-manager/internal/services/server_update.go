package services

import (
	"context"
	"time"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// GetUpdateStatus получает статус обновления
func (s *ServerService) GetUpdateStatus(ctx context.Context, serverID string) (*domain.UpdateStatus, error) {
	if s.updateRepo == nil {
		return &domain.UpdateStatus{
			ServerID:  serverID,
			Status:    "unknown",
			Progress:  0,
			Message:   "Update repository not initialized",
			StartedAt: time.Now(),
		}, nil
	}
	return s.updateRepo.GetUpdateStatus(ctx, serverID)
}

// StartUpdate запускает обновление
func (s *ServerService) StartUpdate(ctx context.Context, req *domain.UpdateRequest) error {
	// TODO: реализовать запуск обновления
	s.logger.Info("starting update", zap.String("server_id", req.ServerID), zap.String("version", req.Version))
	return nil
}

// CancelUpdate отменяет обновление
func (s *ServerService) CancelUpdate(ctx context.Context, serverID string) error {
	// TODO: реализовать отмену обновления
	s.logger.Info("canceling update", zap.String("server_id", serverID))
	return nil
}
