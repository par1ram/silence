package services

import (
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
)

// HealthService сервис для health check
type HealthService struct {
	service string
	version string
}

// NewHealthService создает новый health service
func NewHealthService(service, version string) *HealthService {
	return &HealthService{
		service: service,
		version: version,
	}
}

// GetHealth возвращает статус здоровья сервиса
func (h *HealthService) GetHealth() *domain.HealthStatus {
	return &domain.HealthStatus{
		Status:    "ok",
		Service:   h.service,
		Version:   h.version,
		Timestamp: time.Now(),
	}
}
