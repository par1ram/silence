package ports

import "github.com/par1ram/silence/api/gateway/internal/domain"

// HealthService интерфейс для health check сервиса
type HealthService interface {
	GetHealth() *domain.HealthStatus
}
