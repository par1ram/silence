package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/par1ram/silence/api/gateway/internal/ports"
	"go.uber.org/zap"
)

// ===== Типы запросов/ответов =====

type ConnectRequest struct {
	BypassMethod string `json:"bypass_method"`
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

type ConnectResponse struct {
	BypassID   string    `json:"bypass_id"`
	BypassPort int       `json:"bypass_port"`
	VPNTunnel  string    `json:"vpn_tunnel"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type RateLimitWhitelistRequest struct {
	IP string `json:"ip"`
}

type RateLimitWhitelistResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	IP      string `json:"ip,omitempty"`
}

type RateLimitStatsResponse struct {
	TotalRequests       int64   `json:"total_requests"`
	BlockedRequests     int64   `json:"blocked_requests"`
	WhitelistedRequests int64   `json:"whitelisted_requests"`
	BlockRatePercent    float64 `json:"block_rate_percent"`
}

// ===== Handlers =====

type Handlers struct {
	healthService ports.HealthService
	proxyService  ports.ProxyService
	rateLimiter   *RateLimiter
	logger        *zap.Logger
}

func NewHandlers(healthService ports.HealthService, proxyService ports.ProxyService, rateLimiter *RateLimiter, logger *zap.Logger) *Handlers {
	return &Handlers{
		healthService: healthService,
		proxyService:  proxyService,
		rateLimiter:   rateLimiter,
		logger:        logger,
	}
}

// ===== Вспомогательные методы =====

func writeJSON(w http.ResponseWriter, status int, v interface{}, logger *zap.Logger, logMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil && logger != nil {
		logger.Error(logMsg, zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

// ===== Health =====

func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.healthService.GetHealth(), h.logger, "failed to encode health response")
}

func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"message": "Silence VPN Gateway Service",
		"version": "1.0.0",
	}
	writeJSON(w, http.StatusOK, resp, h.logger, "failed to encode root response")
}

// ===== Прокси =====

func (h *Handlers) AuthHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToAuth(w, r)
}

func (h *Handlers) VPNHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToVPNCore(w, r)
}

func (h *Handlers) DPIHandler(w http.ResponseWriter, r *http.Request) {
	h.proxyService.ProxyToDPIBypass(w, r)
}

// ===== VPN + Bypass Connect =====

func (h *Handlers) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

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
		writeError(w, http.StatusInternalServerError, "Failed to create bypass")
		return
	}

	if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
		h.logger.Error("failed to start bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start bypass")
		return
	}

	vpnReq := map[string]interface{}{
		"name":          req.VPNConfig.Name,
		"listen_port":   req.VPNConfig.ListenPort,
		"mtu":           req.VPNConfig.MTU,
		"auto_recovery": req.VPNConfig.AutoRecovery,
	}
	vpnResp, err := h.proxyService.CreateVPNTunnel(r.Context(), vpnReq)
	if err != nil {
		h.logger.Error("failed to create VPN tunnel", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create VPN tunnel")
		return
	}

	response := ConnectResponse{
		BypassID:   bypassResp["id"].(string),
		BypassPort: req.BypassConfig.LocalPort,
		VPNTunnel:  vpnResp["id"].(string),
		Status:     "connected",
		CreatedAt:  time.Now(),
	}
	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode connect response")
}

// ===== Rate Limiting API =====

func (h *Handlers) RateLimitWhitelistAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req RateLimitWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.IP == "" {
		writeError(w, http.StatusBadRequest, "IP address is required")
		return
	}
	h.rateLimiter.AddToWhitelist(req.IP)
	resp := RateLimitWhitelistResponse{Success: true, Message: "IP added to whitelist successfully", IP: req.IP}
	writeJSON(w, http.StatusOK, resp, h.logger, "failed to encode whitelist response")
}

func (h *Handlers) RateLimitWhitelistRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req RateLimitWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.IP == "" {
		writeError(w, http.StatusBadRequest, "IP address is required")
		return
	}
	h.rateLimiter.RemoveFromWhitelist(req.IP)
	resp := RateLimitWhitelistResponse{Success: true, Message: "IP removed from whitelist successfully", IP: req.IP}
	writeJSON(w, http.StatusOK, resp, h.logger, "failed to encode whitelist response")
}

func (h *Handlers) RateLimitStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	h.rateLimiter.statsMu.RLock()
	stats := h.rateLimiter.stats
	h.rateLimiter.statsMu.RUnlock()
	var blockRate float64
	if stats.requests > 0 {
		blockRate = float64(stats.blocked) / float64(stats.requests) * 100
	}
	resp := RateLimitStatsResponse{
		TotalRequests:       stats.requests,
		BlockedRequests:     stats.blocked,
		WhitelistedRequests: stats.whitelisted,
		BlockRatePercent:    blockRate,
	}
	writeJSON(w, http.StatusOK, resp, h.logger, "failed to encode stats response")
}
