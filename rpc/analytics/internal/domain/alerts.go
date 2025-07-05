package domain

import (
	"time"
)

// AlertSeverity уровень важности уведомления
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus статус уведомления
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusInactive     AlertStatus = "inactive"
	AlertStatusTriggered    AlertStatus = "triggered"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// AlertRule правило уведомления
type AlertRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Condition   string        `json:"condition"` // выражение для оценки
	Severity    AlertSeverity `json:"severity"`
	Message     string        `json:"message"`
	Status      AlertStatus   `json:"status"`
	Enabled     bool          `json:"enabled"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Alert уведомление
type Alert struct {
	ID             string        `json:"id"`
	RuleID         string        `json:"rule_id"`
	Severity       AlertSeverity `json:"severity"`
	Message        string        `json:"message"`
	Status         AlertStatus   `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time    `json:"resolved_at,omitempty"`
	MetricValue    float64       `json:"metric_value,omitempty"`
	ServerID       string        `json:"server_id,omitempty"`
	UserID         string        `json:"user_id,omitempty"`
}
