package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/server-manager/internal/adapters/docker"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"go.uber.org/zap"
)

// ServerService реализация сервиса управления серверами
type ServerService struct {
	serverRepo    ports.ServerRepository
	statsRepo     ports.StatsRepository
	healthRepo    ports.HealthRepository
	scalingRepo   ports.ScalingRepository
	backupRepo    ports.BackupRepository
	updateRepo    ports.UpdateRepository
	dockerAdapter *docker.DockerAdapter
	logger        *zap.Logger
	mutex         sync.RWMutex
}

// NewServerService создает новый сервис управления серверами
func NewServerService(
	serverRepo ports.ServerRepository,
	statsRepo ports.StatsRepository,
	healthRepo ports.HealthRepository,
	scalingRepo ports.ScalingRepository,
	backupRepo ports.BackupRepository,
	updateRepo ports.UpdateRepository,
	dockerAdapter *docker.DockerAdapter,
	logger *zap.Logger,
) ports.ServerService {
	return &ServerService{
		serverRepo:    serverRepo,
		statsRepo:     statsRepo,
		healthRepo:    healthRepo,
		scalingRepo:   scalingRepo,
		backupRepo:    backupRepo,
		updateRepo:    updateRepo,
		dockerAdapter: dockerAdapter,
		logger:        logger,
	}
}

// CreateServer создает новый сервер
func (s *ServerService) CreateServer(ctx context.Context, req *domain.CreateServerRequest) (*domain.Server, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Создаем сервер в базе данных
	server := &domain.Server{
		Name:   req.Name,
		Type:   req.Type,
		Status: domain.ServerStatusCreating,
		Region: req.Region,
	}

	if err := s.serverRepo.Create(ctx, server); err != nil {
		return nil, fmt.Errorf("failed to create server in database: %w", err)
	}

	// Определяем Docker образ в зависимости от типа сервера
	image := s.getImageForServerType(req.Type)
	containerName := fmt.Sprintf("silence-%s-%s", req.Type, server.ID)

	// Создаем контейнер
	containerID, err := s.dockerAdapter.CreateContainer(ctx, containerName, image, req.Config)
	if err != nil {
		// Обновляем статус на ошибку
		server.Status = domain.ServerStatusError
		s.serverRepo.Update(ctx, server)
		return nil, fmt.Errorf("failed to create docker container: %w", err)
	}

	// Запускаем контейнер
	if err := s.dockerAdapter.StartContainer(ctx, containerID); err != nil {
		server.Status = domain.ServerStatusError
		s.serverRepo.Update(ctx, server)
		return nil, fmt.Errorf("failed to start docker container: %w", err)
	}

	// Обновляем статус на запущенный
	server.Status = domain.ServerStatusRunning
	if err := s.serverRepo.Update(ctx, server); err != nil {
		s.logger.Error("failed to update server status", zap.String("server_id", server.ID), zap.Error(err))
	}

	s.logger.Info("server created successfully",
		zap.String("server_id", server.ID),
		zap.String("name", server.Name),
		zap.String("type", string(server.Type)),
		zap.String("container_id", containerID))

	return server, nil
}

// GetServer получает сервер по ID
func (s *ServerService) GetServer(ctx context.Context, id string) (*domain.Server, error) {
	return s.serverRepo.GetByID(ctx, id)
}

// ListServers получает список серверов с фильтрами
func (s *ServerService) ListServers(ctx context.Context, filters map[string]interface{}) ([]*domain.Server, error) {
	return s.serverRepo.List(ctx, filters)
}

// UpdateServer обновляет сервер
func (s *ServerService) UpdateServer(ctx context.Context, id string, req *domain.UpdateServerRequest) (*domain.Server, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Status != nil {
		server.Status = *req.Status
	}

	// Обновляем в базе данных
	if err := s.serverRepo.Update(ctx, server); err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	return server, nil
}

// DeleteServer удаляет сервер
func (s *ServerService) DeleteServer(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Останавливаем контейнер если он запущен
	if server.Status == domain.ServerStatusRunning {
		// TODO: получить container ID из метаданных сервера
		// if err := s.dockerAdapter.StopContainer(ctx, containerID, nil); err != nil {
		//     s.logger.Error("failed to stop container", zap.String("server_id", id), zap.Error(err))
		// }
	}

	// Удаляем из базы данных (soft delete)
	if err := s.serverRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	s.logger.Info("server deleted", zap.String("server_id", id))
	return nil
}

// StartServer запускает сервер
func (s *ServerService) StartServer(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if server.Status == domain.ServerStatusRunning {
		return fmt.Errorf("server is already running")
	}

	// TODO: запустить контейнер
	// containerID := getContainerID(server)
	// if err := s.dockerAdapter.StartContainer(ctx, containerID); err != nil {
	//     return fmt.Errorf("failed to start container: %w", err)
	// }

	server.Status = domain.ServerStatusRunning
	if err := s.serverRepo.Update(ctx, server); err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	s.logger.Info("server started", zap.String("server_id", id))
	return nil
}

// StopServer останавливает сервер
func (s *ServerService) StopServer(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if server.Status == domain.ServerStatusStopped {
		return fmt.Errorf("server is already stopped")
	}

	// TODO: остановить контейнер
	// containerID := getContainerID(server)
	// if err := s.dockerAdapter.StopContainer(ctx, containerID, nil); err != nil {
	//     return fmt.Errorf("failed to stop container: %w", err)
	// }

	server.Status = domain.ServerStatusStopped
	if err := s.serverRepo.Update(ctx, server); err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	s.logger.Info("server stopped", zap.String("server_id", id))
	return nil
}

// RestartServer перезапускает сервер
func (s *ServerService) RestartServer(ctx context.Context, id string) error {
	if err := s.StopServer(ctx, id); err != nil {
		return err
	}

	time.Sleep(2 * time.Second) // Небольшая пауза

	return s.StartServer(ctx, id)
}

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

// GetScalingPolicies получает политики масштабирования
func (s *ServerService) GetScalingPolicies(ctx context.Context) ([]*domain.ScalingPolicy, error) {
	if s.scalingRepo == nil {
		return []*domain.ScalingPolicy{}, nil
	}
	return s.scalingRepo.ListPolicies(ctx)
}

// CreateScalingPolicy создает политику масштабирования
func (s *ServerService) CreateScalingPolicy(ctx context.Context, policy *domain.ScalingPolicy) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	policy.ID = uuid.New().String()
	return s.scalingRepo.SavePolicy(ctx, policy)
}

// UpdateScalingPolicy обновляет политику масштабирования
func (s *ServerService) UpdateScalingPolicy(ctx context.Context, id string, policy *domain.ScalingPolicy) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	policy.ID = id
	return s.scalingRepo.UpdatePolicy(ctx, policy)
}

// DeleteScalingPolicy удаляет политику масштабирования
func (s *ServerService) DeleteScalingPolicy(ctx context.Context, id string) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	return s.scalingRepo.DeletePolicy(ctx, id)
}

// EvaluateScaling оценивает необходимость масштабирования
func (s *ServerService) EvaluateScaling(ctx context.Context) error {
	// TODO: реализовать логику масштабирования
	s.logger.Info("evaluating scaling policies")
	return nil
}

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

// getImageForServerType возвращает Docker образ для типа сервера
func (s *ServerService) getImageForServerType(serverType domain.ServerType) string {
	switch serverType {
	case domain.ServerTypeVPN:
		return "silence/vpn-core:latest"
	case domain.ServerTypeDPI:
		return "silence/dpi-bypass:latest"
	case domain.ServerTypeGateway:
		return "silence/gateway:latest"
	case domain.ServerTypeAnalytics:
		return "silence/analytics:latest"
	default:
		return "silence/base:latest"
	}
}
