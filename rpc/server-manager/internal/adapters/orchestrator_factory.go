package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/server-manager/internal/adapters/docker"
	"github.com/par1ram/silence/rpc/server-manager/internal/adapters/kubernetes"
	"github.com/par1ram/silence/rpc/server-manager/internal/config"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"go.uber.org/zap"
)

// OrchestratorFactory фабрика для создания оркестраторов
type OrchestratorFactory struct {
	config *config.Config
	logger *zap.Logger
}

// NewOrchestratorFactory создает новую фабрику оркестраторов
func NewOrchestratorFactory(config *config.Config, logger *zap.Logger) *OrchestratorFactory {
	return &OrchestratorFactory{
		config: config,
		logger: logger,
	}
}

// CreateOrchestrator создает оркестратор в зависимости от конфигурации
func (f *OrchestratorFactory) CreateOrchestrator() (ports.Orchestrator, error) {
	switch f.config.Orchestrator.Type {
	case "docker":
		return f.createDockerOrchestrator()
	case "kubernetes":
		return f.createKubernetesOrchestrator()
	default:
		return nil, fmt.Errorf("unsupported orchestrator type: %s", f.config.Orchestrator.Type)
	}
}

// createDockerOrchestrator создает Docker оркестратор
func (f *OrchestratorFactory) createDockerOrchestrator() (ports.Orchestrator, error) {
	dockerAdapter, err := docker.NewDockerAdapter(
		f.config.Docker.Host,
		f.config.Docker.APIVersion,
		f.config.Docker.Timeout,
		f.logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker adapter: %w", err)
	}

	return &DockerOrchestrator{
		adapter: dockerAdapter,
		logger:  f.logger,
	}, nil
}

// createKubernetesOrchestrator создает Kubernetes оркестратор
func (f *OrchestratorFactory) createKubernetesOrchestrator() (ports.Orchestrator, error) {
	k8sAdapter, err := kubernetes.NewKubernetesAdapter(
		f.config.Orchestrator.Kubeconfig,
		f.config.Orchestrator.Namespace,
		f.logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes adapter: %w", err)
	}

	return &KubernetesOrchestrator{
		adapter: k8sAdapter,
		logger:  f.logger,
	}, nil
}

// DockerOrchestrator обертка для Docker адаптера
type DockerOrchestrator struct {
	adapter *docker.DockerAdapter
	logger  *zap.Logger
}

// CreateServer создает сервер в Docker
func (d *DockerOrchestrator) CreateServer(ctx context.Context, server *domain.Server) error {
	// Создаем конфигурацию контейнера
	config := map[string]interface{}{
		"environment": map[string]interface{}{
			"SERVER_ID":   server.ID,
			"SERVER_TYPE": string(server.Type),
			"REGION":      server.Region,
		},
	}

	containerID, err := d.adapter.CreateContainer(ctx, server.Name, "silence/vpn-core:latest", config)
	if err != nil {
		return err
	}

	// Запускаем контейнер
	return d.adapter.StartContainer(ctx, containerID)
}

// StartServer запускает сервер в Docker
func (d *DockerOrchestrator) StartServer(ctx context.Context, serverID string) error {
	// В Docker нужно найти контейнер по имени и запустить его
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+serverID {
				return d.adapter.StartContainer(ctx, container.ID)
			}
		}
	}

	return fmt.Errorf("container not found: %s", serverID)
}

// StopServer останавливает сервер в Docker
func (d *DockerOrchestrator) StopServer(ctx context.Context, serverID string) error {
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+serverID {
				timeout := 30 * time.Second
				return d.adapter.StopContainer(ctx, container.ID, &timeout)
			}
		}
	}

	return fmt.Errorf("container not found: %s", serverID)
}

// DeleteServer удаляет сервер в Docker
func (d *DockerOrchestrator) DeleteServer(ctx context.Context, serverID string) error {
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+serverID {
				return d.adapter.RemoveContainer(ctx, container.ID, true)
			}
		}
	}

	return fmt.Errorf("container not found: %s", serverID)
}

// GetServerStats получает статистику сервера из Docker
func (d *DockerOrchestrator) GetServerStats(ctx context.Context, serverID string) (*domain.ServerStats, error) {
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+serverID {
				return d.adapter.GetContainerStats(ctx, container.ID)
			}
		}
	}

	return &domain.ServerStats{
		ServerID:     serverID,
		CPUUsage:     0.0,
		MemoryUsage:  0.0,
		StorageUsage: 0.0,
		NetworkIn:    0,
		NetworkOut:   0,
		Uptime:       0,
		RequestCount: 0,
		ResponseTime: 0.0,
		ErrorRate:    0.0,
		Timestamp:    time.Now(),
	}, nil
}

// GetServerHealth получает здоровье сервера из Docker
func (d *DockerOrchestrator) GetServerHealth(ctx context.Context, serverID string) (*domain.ServerHealth, error) {
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+serverID {
				return d.adapter.GetContainerHealth(ctx, container.ID)
			}
		}
	}

	return &domain.ServerHealth{
		ServerID:    serverID,
		Status:      domain.ServerStatusError,
		Message:     "Container not found",
		LastCheckAt: time.Now(),
		Checks:      []map[string]interface{}{},
	}, nil
}

// ScaleServer масштабирует сервер в Docker (не поддерживается)
func (d *DockerOrchestrator) ScaleServer(ctx context.Context, serverID string, replicas int32) error {
	return fmt.Errorf("scaling not supported in Docker mode")
}

// ListServers получает список серверов из Docker
func (d *DockerOrchestrator) ListServers(ctx context.Context) ([]*domain.Server, error) {
	containers, err := d.adapter.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	var servers []*domain.Server
	for _, container := range containers {
		// Фильтруем только наши контейнеры
		if len(container.Names) > 0 {
			server := &domain.Server{
				ID:        container.Names[0][1:], // Убираем "/" в начале
				Name:      container.Names[0][1:],
				Status:    domain.ServerStatus(container.State),
				CreatedAt: time.Unix(container.Created, 0),
				UpdatedAt: time.Now(),
			}
			servers = append(servers, server)
		}
	}

	return servers, nil
}

// KubernetesOrchestrator обертка для Kubernetes адаптера
type KubernetesOrchestrator struct {
	adapter *kubernetes.KubernetesAdapter
	logger  *zap.Logger
}

// CreateServer создает сервер в Kubernetes
func (k *KubernetesOrchestrator) CreateServer(ctx context.Context, server *domain.Server) error {
	return k.adapter.CreateServer(ctx, server)
}

// StartServer запускает сервер в Kubernetes
func (k *KubernetesOrchestrator) StartServer(ctx context.Context, serverID string) error {
	return k.adapter.StartServer(ctx, serverID)
}

// StopServer останавливает сервер в Kubernetes
func (k *KubernetesOrchestrator) StopServer(ctx context.Context, serverID string) error {
	return k.adapter.StopServer(ctx, serverID)
}

// DeleteServer удаляет сервер в Kubernetes
func (k *KubernetesOrchestrator) DeleteServer(ctx context.Context, serverID string) error {
	return k.adapter.DeleteServer(ctx, serverID)
}

// GetServerStats получает статистику сервера из Kubernetes
func (k *KubernetesOrchestrator) GetServerStats(ctx context.Context, serverID string) (*domain.ServerStats, error) {
	return k.adapter.GetServerStats(ctx, serverID)
}

// GetServerHealth получает здоровье сервера из Kubernetes
func (k *KubernetesOrchestrator) GetServerHealth(ctx context.Context, serverID string) (*domain.ServerHealth, error) {
	return k.adapter.GetServerHealth(ctx, serverID)
}

// ScaleServer масштабирует сервер в Kubernetes
func (k *KubernetesOrchestrator) ScaleServer(ctx context.Context, serverID string, replicas int32) error {
	return k.adapter.ScaleServer(ctx, serverID, replicas)
}

// ListServers получает список серверов из Kubernetes
func (k *KubernetesOrchestrator) ListServers(ctx context.Context) ([]*domain.Server, error) {
	return k.adapter.ListServers(ctx)
}
