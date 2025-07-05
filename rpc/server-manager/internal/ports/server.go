package ports

import (
	"context"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
)

// ServerService интерфейс для управления серверами
type ServerService interface {
	// CRUD операции
	CreateServer(ctx context.Context, req *domain.CreateServerRequest) (*domain.Server, error)
	GetServer(ctx context.Context, id string) (*domain.Server, error)
	ListServers(ctx context.Context, filters map[string]interface{}) ([]*domain.Server, error)
	UpdateServer(ctx context.Context, id string, req *domain.UpdateServerRequest) (*domain.Server, error)
	DeleteServer(ctx context.Context, id string) error

	// Управление жизненным циклом
	StartServer(ctx context.Context, id string) error
	StopServer(ctx context.Context, id string) error
	RestartServer(ctx context.Context, id string) error

	// Мониторинг
	GetServerStats(ctx context.Context, id string) (*domain.ServerStats, error)
	GetServerHealth(ctx context.Context, id string) (*domain.ServerHealth, error)
	GetAllServersHealth(ctx context.Context) ([]*domain.ServerHealth, error)

	// Масштабирование
	GetScalingPolicies(ctx context.Context) ([]*domain.ScalingPolicy, error)
	CreateScalingPolicy(ctx context.Context, policy *domain.ScalingPolicy) error
	UpdateScalingPolicy(ctx context.Context, id string, policy *domain.ScalingPolicy) error
	DeleteScalingPolicy(ctx context.Context, id string) error
	EvaluateScaling(ctx context.Context) error

	// Резервное копирование
	GetBackupConfigs(ctx context.Context) ([]*domain.BackupConfig, error)
	CreateBackupConfig(ctx context.Context, config *domain.BackupConfig) error
	UpdateBackupConfig(ctx context.Context, id string, config *domain.BackupConfig) error
	DeleteBackupConfig(ctx context.Context, id string) error
	CreateBackup(ctx context.Context, serverID string) error
	RestoreBackup(ctx context.Context, serverID, backupID string) error

	// Обновления
	GetUpdateStatus(ctx context.Context, serverID string) (*domain.UpdateStatus, error)
	StartUpdate(ctx context.Context, req *domain.UpdateRequest) error
	CancelUpdate(ctx context.Context, serverID string) error
}

// ServerRepository интерфейс для работы с базой данных серверов
type ServerRepository interface {
	Create(ctx context.Context, server *domain.Server) error
	GetByID(ctx context.Context, id string) (*domain.Server, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.Server, error)
	Update(ctx context.Context, server *domain.Server) error
	Delete(ctx context.Context, id string) error
	GetByType(ctx context.Context, serverType domain.ServerType) ([]*domain.Server, error)
	GetByRegion(ctx context.Context, region string) ([]*domain.Server, error)
	GetByStatus(ctx context.Context, status domain.ServerStatus) ([]*domain.Server, error)
}

// StatsRepository интерфейс для работы со статистикой
type StatsRepository interface {
	SaveStats(ctx context.Context, stats *domain.ServerStats) error
	GetStats(ctx context.Context, serverID string, limit int) ([]*domain.ServerStats, error)
	GetLatestStats(ctx context.Context, serverID string) (*domain.ServerStats, error)
	GetAggregatedStats(ctx context.Context, serverID string, period string) (*domain.ServerStats, error)
}

// HealthRepository интерфейс для работы с данными о здоровье
type HealthRepository interface {
	SaveHealth(ctx context.Context, health *domain.ServerHealth) error
	GetHealth(ctx context.Context, serverID string) (*domain.ServerHealth, error)
	GetAllHealth(ctx context.Context) ([]*domain.ServerHealth, error)
	GetHealthHistory(ctx context.Context, serverID string, limit int) ([]*domain.ServerHealth, error)
}

// ScalingRepository интерфейс для работы с политиками масштабирования
type ScalingRepository interface {
	SavePolicy(ctx context.Context, policy *domain.ScalingPolicy) error
	GetPolicy(ctx context.Context, id string) (*domain.ScalingPolicy, error)
	ListPolicies(ctx context.Context) ([]*domain.ScalingPolicy, error)
	UpdatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error
	DeletePolicy(ctx context.Context, id string) error
}

// BackupRepository интерфейс для работы с резервными копиями
type BackupRepository interface {
	SaveConfig(ctx context.Context, config *domain.BackupConfig) error
	GetConfig(ctx context.Context, id string) (*domain.BackupConfig, error)
	ListConfigs(ctx context.Context) ([]*domain.BackupConfig, error)
	UpdateConfig(ctx context.Context, config *domain.BackupConfig) error
	DeleteConfig(ctx context.Context, id string) error
	SaveBackup(ctx context.Context, serverID, backupID string, metadata map[string]interface{}) error
	GetBackups(ctx context.Context, serverID string) ([]map[string]interface{}, error)
	DeleteBackup(ctx context.Context, backupID string) error
}

// UpdateRepository интерфейс для работы с обновлениями
type UpdateRepository interface {
	SaveUpdateStatus(ctx context.Context, status *domain.UpdateStatus) error
	GetUpdateStatus(ctx context.Context, serverID string) (*domain.UpdateStatus, error)
	UpdateProgress(ctx context.Context, serverID string, progress int, message string) error
	CompleteUpdate(ctx context.Context, serverID string, success bool, message string) error
}
