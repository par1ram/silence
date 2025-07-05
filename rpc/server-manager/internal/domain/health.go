package domain

import "time"

// HealthStatus статус здоровья сервиса
type HealthStatus struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// NewHealthStatus создает новый статус здоровья
func NewHealthStatus(service, version string) *HealthStatus {
	return &HealthStatus{
		Status:    "ok",
		Service:   service,
		Version:   version,
		Timestamp: time.Now(),
	}
}
