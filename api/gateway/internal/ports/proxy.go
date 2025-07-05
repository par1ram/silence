package ports

import (
	"context"
	"net/http"
)

// ProxyService интерфейс для проксирования запросов
type ProxyService interface {
	ProxyToAuth(w http.ResponseWriter, r *http.Request)
	ProxyToVPNCore(w http.ResponseWriter, r *http.Request)
	HealthCheck(ctx context.Context) error
}
