package domain

import (
	"time"
)

// ServerStatus статус сервера
type ServerStatus string

const (
	ServerStatusCreating ServerStatus = "creating"
	ServerStatusRunning  ServerStatus = "running"
	ServerStatusStopped  ServerStatus = "stopped"
	ServerStatusError    ServerStatus = "error"
	ServerStatusDeleting ServerStatus = "deleting"
)

// ServerType тип сервера
type ServerType string

const (
	ServerTypeVPN       ServerType = "vpn"
	ServerTypeDPI       ServerType = "dpi"
	ServerTypeGateway   ServerType = "gateway"
	ServerTypeAnalytics ServerType = "analytics"
)

// Server модель сервера
type Server struct {
	ID        string       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Type      ServerType   `json:"type" db:"type"`
	Status    ServerStatus `json:"status" db:"status"`
	Region    string       `json:"region" db:"region"`
	IP        string       `json:"ip" db:"ip"`
	Port      int          `json:"port" db:"port"`
	CPU       float64      `json:"cpu" db:"cpu"`
	Memory    float64      `json:"memory" db:"memory"`
	Disk      float64      `json:"disk" db:"disk"`
	Network   float64      `json:"network" db:"network"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateServerRequest запрос на создание сервера
type CreateServerRequest struct {
	Name   string            `json:"name" validate:"required"`
	Type   ServerType        `json:"type" validate:"required"`
	Region string            `json:"region" validate:"required"`
	Config map[string]string `json:"config,omitempty"`
}

// UpdateServerRequest запрос на обновление сервера
type UpdateServerRequest struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Type   ServerType        `json:"type"`
	Region string            `json:"region"`
	Config map[string]string `json:"config,omitempty"`
}

// ServerStatsOld статистика сервера (старая версия)
type ServerStatsOld struct {
	ServerID    string    `json:"server_id"`
	CPU         float64   `json:"cpu"`
	Memory      float64   `json:"memory"`
	Disk        float64   `json:"disk"`
	Network     float64   `json:"network"`
	Connections int       `json:"connections"`
	Timestamp   time.Time `json:"timestamp"`
}

// ServerHealthOld здоровье сервера (старая версия)
type ServerHealthOld struct {
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ScalingPolicy политика масштабирования
type ScalingPolicy struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	MinServers        int           `json:"min_servers"`
	MaxServers        int           `json:"max_servers"`
	CPUThreshold      float64       `json:"cpu_threshold"`
	MemoryThreshold   float64       `json:"memory_threshold"`
	ScaleUpCooldown   time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown time.Duration `json:"scale_down_cooldown"`
	Enabled           bool          `json:"enabled"`
}

// BackupConfig конфигурация резервного копирования
type BackupConfig struct {
	ID          string     `json:"id"`
	ServerID    string     `json:"server_id"`
	Schedule    string     `json:"schedule"`  // cron expression
	Retention   int        `json:"retention"` // days
	Type        string     `json:"type"`      // full, incremental
	Destination string     `json:"destination"`
	Enabled     bool       `json:"enabled"`
	LastBackup  *time.Time `json:"last_backup,omitempty"`
	NextBackup  *time.Time `json:"next_backup,omitempty"`
}

// UpdateRequest запрос на обновление
type UpdateRequest struct {
	ServerID string `json:"server_id"`
	Version  string `json:"version"`
	Force    bool   `json:"force"`
}

// UpdateStatus статус обновления
type UpdateStatus struct {
	ServerID    string     `json:"server_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Message     string     `json:"message"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ServerFilters фильтры для списка серверов
type ServerFilters struct {
	Type   ServerType   `json:"type"`
	Region string       `json:"region"`
	Status ServerStatus `json:"status"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// ServerMonitorEvent событие мониторинга сервера
type ServerMonitorEvent struct {
	ServerID  string    `json:"server_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ServerStats статистика сервера для gRPC
type ServerStats struct {
	ServerID     string    `json:"server_id"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	StorageUsage float64   `json:"storage_usage"`
	NetworkIn    int64     `json:"network_in"`
	NetworkOut   int64     `json:"network_out"`
	Uptime       int64     `json:"uptime"`
	RequestCount int64     `json:"request_count"`
	ResponseTime float64   `json:"response_time"`
	ErrorRate    float64   `json:"error_rate"`
	Timestamp    time.Time `json:"timestamp"`
}

// ServerHealth здоровье сервера для gRPC
type ServerHealth struct {
	ServerID    string                   `json:"server_id"`
	Status      ServerStatus             `json:"status"`
	Message     string                   `json:"message"`
	LastCheckAt time.Time                `json:"last_check_at"`
	Checks      []map[string]interface{} `json:"checks"`
}

// ScaleServerRequest запрос на масштабирование сервера
type ScaleServerRequest struct {
	ServerID string  `json:"server_id"`
	CPU      float64 `json:"cpu"`
	Memory   float64 `json:"memory"`
	Storage  float64 `json:"storage"`
	Replicas int     `json:"replicas"`
}

// CreateBackupRequest запрос на создание резервной копии
type CreateBackupRequest struct {
	ServerID    string `json:"server_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RestoreBackupRequest запрос на восстановление из резервной копии
type RestoreBackupRequest struct {
	BackupID string `json:"backup_id"`
	ServerID string `json:"server_id"`
}

// UpdateServerSoftwareRequest запрос на обновление ПО сервера
type UpdateServerSoftwareRequest struct {
	ServerID string `json:"server_id"`
	Version  string `json:"version"`
	Package  string `json:"package"`
	Force    bool   `json:"force"`
}

// Backup резервная копия
type Backup struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
