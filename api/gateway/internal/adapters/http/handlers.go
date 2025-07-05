package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/par1ram/silence/api/gateway/internal/ports"
	"go.uber.org/zap"
)

// ConnectRequest запрос на создание VPN-соединения с обфускацией
type ConnectRequest struct {
	BypassMethod string `json:"bypass_method"` // shadowsocks, v2ray, obfs4, custom
	BypassConfig struct {
		LocalPort  int    `json:"local_port"`
		RemoteHost string `json:"remote_host"`
		RemotePort int    `json:"remote_port"`
		Password   string `json:"password,omitempty"`
		Encryption string `json:"encryption"`
	} `json:"bypass_config"`
	VPNConfig struct {
		Name         string `json:"name"`
		ListenPort   int    `json:"listen_port"`
		MTU          int    `json:"mtu"`
		AutoRecovery bool   `json:"auto_recovery"`
	} `json:"vpn_config"`
}

// ConnectResponse ответ с информацией о созданном соединении
type ConnectResponse struct {
	BypassID   string    `json:"bypass_id"`
	BypassPort int       `json:"bypass_port"`
	VPNTunnel  string    `json:"vpn_tunnel"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

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

// DPIHandler проксирует запросы к DPI Bypass сервису
func (h *Handlers) DPIHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToDPIBypass(w, r)
}

// ConnectHandler создает VPN-соединение с обфускацией
func (h *Handlers) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаем bypass-конфигурацию
	bypassReq := map[string]interface{}{
		"name":        "vpn-bypass-" + time.Now().Format("20060102150405"),
		"method":      req.BypassMethod,
		"local_port":  req.BypassConfig.LocalPort,
		"remote_host": req.BypassConfig.RemoteHost,
		"remote_port": req.BypassConfig.RemotePort,
		"password":    req.BypassConfig.Password,
		"encryption":  req.BypassConfig.Encryption,
	}

	bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
	if err != nil {
		h.logger.Error("failed to create bypass", zap.Error(err))
		http.Error(w, "Failed to create bypass", http.StatusInternalServerError)
		return
	}

	// Запускаем bypass
	if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
		h.logger.Error("failed to start bypass", zap.Error(err))
		http.Error(w, "Failed to start bypass", http.StatusInternalServerError)
		return
	}

	// Создаем VPN-туннель
	vpnReq := map[string]interface{}{
		"name":          req.VPNConfig.Name,
		"listen_port":   req.VPNConfig.ListenPort,
		"mtu":           req.VPNConfig.MTU,
		"auto_recovery": req.VPNConfig.AutoRecovery,
	}

	vpnResp, err := h.proxyService.CreateVPNTunnel(r.Context(), vpnReq)
	if err != nil {
		h.logger.Error("failed to create VPN tunnel", zap.Error(err))
		http.Error(w, "Failed to create VPN tunnel", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := ConnectResponse{
		BypassID:   bypassResp["id"].(string),
		BypassPort: req.BypassConfig.LocalPort,
		VPNTunnel:  vpnResp["id"].(string),
		Status:     "connected",
		CreatedAt:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
