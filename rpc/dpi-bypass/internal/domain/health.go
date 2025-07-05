package domain

import "time"

// HealthStatus статус здоровья сервиса
type HealthStatus struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
