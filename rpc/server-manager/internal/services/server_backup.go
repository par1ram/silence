package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// GetBackupConfigs получает конфигурации резервного копирования
func (s *ServerService) GetBackupConfigs(ctx context.Context) ([]*domain.BackupConfig, error) {
	if s.backupRepo == nil {
		return []*domain.BackupConfig{}, nil
	}
	return s.backupRepo.ListConfigs(ctx)
}

// CreateBackupConfig создает конфигурацию резервного копирования
func (s *ServerService) CreateBackupConfig(ctx context.Context, config *domain.BackupConfig) error {
	if s.backupRepo == nil {
		return fmt.Errorf("backup repository not initialized")
	}
	config.ID = uuid.New().String()
	return s.backupRepo.SaveConfig(ctx, config)
}

// UpdateBackupConfig обновляет конфигурацию резервного копирования
func (s *ServerService) UpdateBackupConfig(ctx context.Context, id string, config *domain.BackupConfig) error {
	if s.backupRepo == nil {
		return fmt.Errorf("backup repository not initialized")
	}
	config.ID = id
	return s.backupRepo.UpdateConfig(ctx, config)
}

// DeleteBackupConfig удаляет конфигурацию резервного копирования
func (s *ServerService) DeleteBackupConfig(ctx context.Context, id string) error {
	if s.backupRepo == nil {
		return fmt.Errorf("backup repository not initialized")
	}
	return s.backupRepo.DeleteConfig(ctx, id)
}

// CreateBackup создает резервную копию
func (s *ServerService) CreateBackup(ctx context.Context, serverID string) error {
	// TODO: реализовать создание резервной копии
	s.logger.Info("creating backup", zap.String("server_id", serverID))
	return nil
}

// RestoreBackup восстанавливает из резервной копии
func (s *ServerService) RestoreBackup(ctx context.Context, serverID, backupID string) error {
	// TODO: реализовать восстановление из резервной копии
	s.logger.Info("restoring backup", zap.String("server_id", serverID), zap.String("backup_id", backupID))
	return nil
}
