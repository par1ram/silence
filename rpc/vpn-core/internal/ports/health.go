package ports

import "github.com/par1ram/silence/rpc/vpn-core/internal/domain"

// HealthService интерфейс для health check
type HealthService interface {
	GetHealth() *domain.HealthStatus
}
