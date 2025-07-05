package services

import (
	"github.com/par1ram/silence/api/gateway/internal/domain"
	"github.com/par1ram/silence/api/gateway/internal/ports"
)

// HealthService реализация health check сервиса
type HealthService struct {
	serviceName string
	version     string
}

// NewHealthService создает новый health check сервис
func NewHealthService(serviceName, version string) ports.HealthService {
	return &HealthService{
		serviceName: serviceName,
		version:     version,
	}
}

// GetHealth возвращает статус здоровья сервиса
func (h *HealthService) GetHealth() *domain.HealthStatus {
	return domain.NewHealthStatus(h.serviceName, h.version)
}
