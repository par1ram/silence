package ports

import "github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"

// HealthService интерфейс для health check
type HealthService interface {
	GetHealth() *domain.HealthStatus
}
