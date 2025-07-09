package services

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	configInterface := make(map[string]interface{})
	for k, v := range req.Config {
		configInterface[k] = v
	}
	containerID, err := s.dockerAdapter.CreateContainer(ctx, containerName, image, configInterface)
	if err != nil {
		// Обновляем статус на ошибку
		server.Status = domain.ServerStatusError
		if err := s.serverRepo.Update(ctx, server); err != nil {
			s.logger.Error("failed to update server status to error", zap.String("server_id", server.ID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to create docker container: %w", err)
	}

	// Запускаем контейнер
	if err := s.dockerAdapter.StartContainer(ctx, containerID); err != nil {
		server.Status = domain.ServerStatusError
		if err := s.serverRepo.Update(ctx, server); err != nil {
			s.logger.Error("failed to update server status to error", zap.String("server_id", server.ID), zap.Error(err))
		}
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
	if req.Name != "" {
		server.Name = req.Name
	}
	// Обновляем конфигурацию
	// TODO: обновить конфигурацию сервера

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
		// TODO: получить container ID из метаданных сервера и остановить контейнер
		s.logger.Info("server is running, container stop not implemented yet", zap.String("server_id", id))
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
