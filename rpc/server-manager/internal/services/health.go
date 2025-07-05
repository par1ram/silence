package services

import (
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
)

// HealthService реализация health check сервиса
type HealthService struct {
	service string
	version string
}

// NewHealthService создает новый health check сервис
func NewHealthService(service, version string) ports.HealthService {
	return &HealthService{
		service: service,
		version: version,
	}
}

// GetHealth возвращает статус здоровья сервиса
func (h *HealthService) GetHealth() *domain.HealthStatus {
	return domain.NewHealthStatus(h.service, h.version)
}
