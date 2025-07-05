package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
)

// GetServerStats получает статистику сервера
func (s *ServerService) GetServerStats(ctx context.Context, id string) (*domain.ServerStats, error) {
	// Проверяем существование сервера
	if _, err := s.serverRepo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	// Получаем последнюю статистику из базы данных
	if s.statsRepo == nil {
		return &domain.ServerStats{
			ServerID:    id,
			CPU:         0.0,
			Memory:      0.0,
			Disk:        0.0,
			Network:     0.0,
			Connections: 0,
			Timestamp:   time.Now(),
		}, nil
	}

	stats, err := s.statsRepo.GetLatestStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get server stats: %w", err)
	}

	return stats, nil
}

// GetServerHealth получает здоровье сервера
func (s *ServerService) GetServerHealth(ctx context.Context, id string) (*domain.ServerHealth, error) {
	// Проверяем существование сервера
	if _, err := s.serverRepo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	// Получаем данные о здоровье из базы данных
	if s.healthRepo == nil {
		return &domain.ServerHealth{
			ServerID:  id,
			Status:    "unknown",
			Message:   "Health repository not initialized",
			Timestamp: time.Now(),
		}, nil
	}

	health, err := s.healthRepo.GetHealth(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get server health: %w", err)
	}

	return health, nil
}

// GetAllServersHealth получает здоровье всех серверов
func (s *ServerService) GetAllServersHealth(ctx context.Context) ([]*domain.ServerHealth, error) {
	if s.healthRepo == nil {
		return []*domain.ServerHealth{}, nil
	}
	return s.healthRepo.GetAllHealth(ctx)
}
