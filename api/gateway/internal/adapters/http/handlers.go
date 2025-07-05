package http

import (
	"encoding/json"
	"net/http"

	"github.com/par1ram/silence/api/gateway/internal/ports"
	"go.uber.org/zap"
)

// Handlers HTTP обработчики
type Handlers struct {
	healthService ports.HealthService
	proxyService  ports.ProxyService
	logger        *zap.Logger
}

// NewHandlers создает новые HTTP обработчики
func NewHandlers(healthService ports.HealthService, proxyService ports.ProxyService, logger *zap.Logger) *Handlers {
	return &Handlers{
		healthService: healthService,
		proxyService:  proxyService,
		logger:        logger,
	}
}

// HealthHandler обработчик для health check
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := h.healthService.GetHealth()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RootHandler обработчик для корневого пути
func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Silence VPN Gateway Service",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode root response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// AuthHandler проксирует запросы к auth сервису
func (h *Handlers) AuthHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToAuth(w, r)
}

// VPNHandler проксирует запросы к VPN Core сервису
func (h *Handlers) VPNHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToVPNCore(w, r)
}
