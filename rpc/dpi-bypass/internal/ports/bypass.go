package ports

import (
	"context"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
)

// BypassService интерфейс для управления DPI bypass
type BypassService interface {
	CreateBypass(ctx context.Context, req *domain.CreateBypassRequest) (*domain.BypassConfig, error)
	GetBypass(ctx context.Context, id string) (*domain.BypassConfig, error)
	ListBypasses(ctx context.Context) ([]*domain.BypassConfig, error)
	StartBypass(ctx context.Context, id string) error
	StopBypass(ctx context.Context, id string) error
	GetBypassStats(ctx context.Context, id string) (*domain.BypassStats, error)
	DeleteBypass(ctx context.Context, id string) error
}

// BypassAdapter интерфейс для адаптеров обфускации
type BypassAdapter interface {
	Start(config *domain.BypassConfig) error
	Stop(id string) error
	GetStats(id string) (*domain.BypassStats, error)
	IsRunning(id string) bool
}
