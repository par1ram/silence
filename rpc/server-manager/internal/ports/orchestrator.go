package ports

import (
	"context"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
)

// Orchestrator интерфейс для управления серверами
type Orchestrator interface {
	// CreateServer создает новый сервер
	CreateServer(ctx context.Context, server *domain.Server) error

	// StartServer запускает сервер
	StartServer(ctx context.Context, serverID string) error

	// StopServer останавливает сервер
	StopServer(ctx context.Context, serverID string) error

	// DeleteServer удаляет сервер
	DeleteServer(ctx context.Context, serverID string) error

	// GetServerStats получает статистику сервера
	GetServerStats(ctx context.Context, serverID string) (*domain.ServerStats, error)

	// GetServerHealth получает здоровье сервера
	GetServerHealth(ctx context.Context, serverID string) (*domain.ServerHealth, error)

	// ScaleServer масштабирует сервер
	ScaleServer(ctx context.Context, serverID string, replicas int32) error

	// ListServers получает список серверов
	ListServers(ctx context.Context) ([]*domain.Server, error)
}
