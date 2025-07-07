package ports

import (
	"context"
	"net/http"
)

// ProxyService интерфейс для проксирования запросов
type ProxyService interface {
	ProxyToAuth(w http.ResponseWriter, r *http.Request)
	ProxyToVPNCore(w http.ResponseWriter, r *http.Request)
	ProxyToDPIBypass(w http.ResponseWriter, r *http.Request)
	ProxyToAnalytics(w http.ResponseWriter, r *http.Request)
	ProxyToServerManager(w http.ResponseWriter, r *http.Request)
	HealthCheck(ctx context.Context) error

	// Методы для интеграции VPN + обфускация
	CreateBypass(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error)
	StartBypass(ctx context.Context, id string) error
	StopBypass(ctx context.Context, id string) error
	CreateVPNTunnel(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error)
	StartVPNTunnel(ctx context.Context, id string) error
	StopVPNTunnel(ctx context.Context, id string) error
}
