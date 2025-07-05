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
	Name   string                 `json:"name" validate:"required"`
	Type   ServerType             `json:"type" validate:"required"`
	Region string                 `json:"region" validate:"required"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// UpdateServerRequest запрос на обновление сервера
type UpdateServerRequest struct {
	Name   *string                `json:"name,omitempty"`
	Status *ServerStatus          `json:"status,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ServerStats статистика сервера
type ServerStats struct {
	ServerID    string    `json:"server_id"`
	CPU         float64   `json:"cpu"`
	Memory      float64   `json:"memory"`
	Disk        float64   `json:"disk"`
	Network     float64   `json:"network"`
	Connections int       `json:"connections"`
	Timestamp   time.Time `json:"timestamp"`
}

// ServerHealth здоровье сервера
type ServerHealth struct {
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
