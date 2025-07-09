package ports

import (
	"context"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
)

// DPIBypassService интерфейс для управления DPI bypass
type DPIBypassService interface {
	// Bypass configuration management
	CreateBypassConfig(ctx context.Context, req *domain.CreateBypassConfigRequest) (*domain.BypassConfig, error)
	GetBypassConfig(ctx context.Context, id string) (*domain.BypassConfig, error)
	ListBypassConfigs(ctx context.Context, filters *domain.BypassConfigFilters) ([]*domain.BypassConfig, int, error)
	UpdateBypassConfig(ctx context.Context, req *domain.UpdateBypassConfigRequest) (*domain.BypassConfig, error)
	DeleteBypassConfig(ctx context.Context, id string) error

	// Bypass operations
	StartBypass(ctx context.Context, req *domain.StartBypassRequest) (*domain.BypassSession, error)
	StopBypass(ctx context.Context, sessionID string) error
	GetBypassStatus(ctx context.Context, sessionID string) (*domain.BypassSessionStatus, error)

	// Statistics and monitoring
	GetBypassStats(ctx context.Context, sessionID string) (*domain.BypassStats, error)
	GetBypassHistory(ctx context.Context, req *domain.BypassHistoryRequest) ([]*domain.BypassHistoryEntry, int, error)

	// Rule management
	AddBypassRule(ctx context.Context, req *domain.AddBypassRuleRequest) (*domain.BypassRule, error)
	UpdateBypassRule(ctx context.Context, req *domain.UpdateBypassRuleRequest) (*domain.BypassRule, error)
	DeleteBypassRule(ctx context.Context, id string) error
	ListBypassRules(ctx context.Context, filters *domain.BypassRuleFilters) ([]*domain.BypassRule, int, error)
}

// BypassAdapter интерфейс для адаптеров обфускации
type BypassAdapter interface {
	Start(config *domain.BypassConfig) error
	Stop(id string) error
	GetStats(id string) (*domain.BypassStats, error)
	IsRunning(id string) bool
}
