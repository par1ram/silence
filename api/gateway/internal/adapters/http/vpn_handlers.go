package http

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ConnectVPNHandler обрабатывает VPN-only подключения
func (h *Handlers) ConnectVPNHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req VPNConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация запроса
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.ListenPort <= 0 {
		writeError(w, http.StatusBadRequest, "Valid listen port is required")
		return
	}

	// Устанавливаем значения по умолчанию
	if req.MTU == 0 {
		req.MTU = 1420
	}

	// Создаем VPN туннель
	vpnReq := map[string]interface{}{
		"name":          req.Name,
		"listen_port":   req.ListenPort,
		"mtu":           req.MTU,
		"auto_recovery": req.AutoRecovery,
	}

	vpnResp, err := h.proxyService.CreateVPNTunnel(r.Context(), vpnReq)
	if err != nil {
		h.logger.Error("failed to create VPN tunnel", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create VPN tunnel")
		return
	}

	// Запускаем туннель
	tunnelID := getStringFromMap(vpnResp, "id")
	if tunnelID == "" {
		writeError(w, http.StatusInternalServerError, "Invalid tunnel response")
		return
	}

	if err := h.proxyService.StartVPNTunnel(r.Context(), tunnelID); err != nil {
		h.logger.Error("failed to start VPN tunnel", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start VPN tunnel")
		return
	}

	// Формируем ответ
	response := VPNConnectResponse{
		TunnelID:   tunnelID,
		TunnelName: req.Name,
		PublicKey:  getStringFromMap(vpnResp, "public_key"),
		Endpoint:   getStringFromMap(vpnResp, "endpoint"),
		Status:     "connected",
		Config:     buildVPNConfig(vpnResp),
		CreatedAt:  time.Now(),
	}

	h.logger.Info("VPN tunnel created successfully",
		zap.String("tunnel_id", tunnelID),
		zap.String("name", req.Name))

	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode VPN connect response")
}

// buildVPNConfig создает конфигурацию VPN из ответа сервиса
func buildVPNConfig(vpnResp map[string]interface{}) VPNConfig {
	return VPNConfig{
		Interface:  getStringFromMap(vpnResp, "interface"),
		PrivateKey: getStringFromMap(vpnResp, "private_key"),
		Address:    getStringFromMap(vpnResp, "address"),
		DNS:        getStringSliceFromMap(vpnResp, "dns"),
		Peer: VPNPeer{
			PublicKey:  getStringFromMap(vpnResp, "peer_public_key"),
			AllowedIPs: getStringSliceFromMap(vpnResp, "allowed_ips"),
			Endpoint:   getStringFromMap(vpnResp, "peer_endpoint"),
		},
	}
}

// Вспомогательные функции для извлечения данных из map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}
