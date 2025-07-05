package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// DockerAdapter адаптер для работы с Docker
type DockerAdapter struct {
	client *client.Client
	logger *zap.Logger
}

// NewDockerAdapter создает новый Docker адаптер
func NewDockerAdapter(host string, apiVersion string, timeout time.Duration, logger *zap.Logger) (*DockerAdapter, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithVersion(apiVersion),
		client.WithTimeout(timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerAdapter{
		client: cli,
		logger: logger,
	}, nil
}

// CreateContainer создает контейнер
func (d *DockerAdapter) CreateContainer(ctx context.Context, name, image string, config map[string]interface{}) (string, error) {
	// Создаем конфигурацию контейнера
	containerConfig := &container.Config{
		Image: image,
		Cmd:   []string{},
		Env:   []string{},
	}

	// Добавляем переменные окружения из конфига
	if env, ok := config["environment"].(map[string]interface{}); ok {
		for k, v := range env {
			containerConfig.Env = append(containerConfig.Env, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// Добавляем команду запуска
	if cmd, ok := config["command"].([]string); ok {
		containerConfig.Cmd = cmd
	}

	// Создаем контейнер
	resp, err := d.client.ContainerCreate(ctx, containerConfig, nil, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	d.logger.Info("container created", zap.String("id", resp.ID), zap.String("name", name))
	return resp.ID, nil
}

// StartContainer запускает контейнер
func (d *DockerAdapter) StartContainer(ctx context.Context, containerID string) error {
	err := d.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Info("container started", zap.String("id", containerID))
	return nil
}

// StopContainer останавливает контейнер
func (d *DockerAdapter) StopContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	if timeout == nil {
		defaultTimeout := 30 * time.Second
		timeout = &defaultTimeout
	}

	seconds := int(timeout.Seconds())
	stopOptions := container.StopOptions{
		Timeout: &seconds,
	}

	err := d.client.ContainerStop(ctx, containerID, stopOptions)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	d.logger.Info("container stopped", zap.String("id", containerID))
	return nil
}

// RemoveContainer удаляет контейнер
func (d *DockerAdapter) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	err := d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	d.logger.Info("container removed", zap.String("id", containerID))
	return nil
}

// GetContainerStats получает статистику контейнера
func (d *DockerAdapter) GetContainerStats(ctx context.Context, containerID string) (*domain.ServerStats, error) {
	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	// Парсим статистику (упрощенно)
	// В реальной реализации нужно парсить JSON из stats.Body
	statsData := &domain.ServerStats{
		ServerID:    containerID,
		CPU:         0.0, // TODO: парсить из stats
		Memory:      0.0, // TODO: парсить из stats
		Disk:        0.0, // TODO: парсить из stats
		Network:     0.0, // TODO: парсить из stats
		Connections: 0,   // TODO: парсить из stats
		Timestamp:   time.Now(),
	}

	return statsData, nil
}

// GetContainerHealth получает здоровье контейнера
func (d *DockerAdapter) GetContainerHealth(ctx context.Context, containerID string) (*domain.ServerHealth, error) {
	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	health := &domain.ServerHealth{
		ServerID:  containerID,
		Status:    string(inspect.State.Status),
		Message:   inspect.State.Error,
		Timestamp: time.Now(),
	}

	return health, nil
}

// ListContainers получает список контейнеров
func (d *DockerAdapter) ListContainers(ctx context.Context) ([]types.Container, error) {
	containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// Close закрывает соединение с Docker
func (d *DockerAdapter) Close() error {
	return d.client.Close()
}
