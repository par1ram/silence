package ports

import "github.com/par1ram/silence/rpc/server-manager/internal/domain"

// HealthService интерфейс для health check
type HealthService interface {
	GetHealth() *domain.HealthStatus
}
