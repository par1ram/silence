package http

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ConnectV2RayHandler обрабатывает V2Ray подключения
func (h *Handlers) ConnectV2RayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req V2RayConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация запроса
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.ServerHost == "" {
		writeError(w, http.StatusBadRequest, "Server host is required")
		return
	}
	if req.UUID == "" {
		writeError(w, http.StatusBadRequest, "UUID is required")
		return
	}

	// Устанавливаем значения по умолчанию
	if req.LocalPort == 0 {
		req.LocalPort = 1080
	}
	if req.Security == "" {
		req.Security = "auto"
	}
	if req.Network == "" {
		req.Network = "tcp"
	}

	// Создаем V2Ray bypass
	bypassReq := map[string]interface{}{
		"name":        req.Name,
		"method":      "v2ray",
		"local_port":  req.LocalPort,
		"remote_host": req.ServerHost,
		"remote_port": req.ServerPort,
		"uuid":        req.UUID,
		"alter_id":    req.AlterID,
		"security":    req.Security,
		"network":     req.Network,
	}

	bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
	if err != nil {
		h.logger.Error("failed to create V2Ray bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create V2Ray bypass")
		return
	}

	// Запускаем bypass
	bypassID := getStringFromMap(bypassResp, "id")
	if bypassID == "" {
		writeError(w, http.StatusInternalServerError, "Invalid bypass response")
		return
	}

	if err := h.proxyService.StartBypass(r.Context(), bypassID); err != nil {
		h.logger.Error("failed to start V2Ray bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start V2Ray bypass")
		return
	}

	// Формируем ответ
	response := V2RayConnectResponse{
		ConnectionID: bypassID,
		LocalPort:    req.LocalPort,
		Status:       "connected",
		Config:       buildV2RayConfig(req),
		CreatedAt:    time.Now(),
	}

	h.logger.Info("V2Ray bypass created successfully",
		zap.String("bypass_id", bypassID),
		zap.String("name", req.Name))

	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode V2Ray connect response")
}

// ConnectObfs4Handler обрабатывает Obfs4 подключения
func (h *Handlers) ConnectObfs4Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req Obfs4ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация запроса
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Bridge == "" {
		writeError(w, http.StatusBadRequest, "Bridge is required")
		return
	}
	if req.Cert == "" {
		writeError(w, http.StatusBadRequest, "Certificate is required")
		return
	}

	// Устанавливаем значения по умолчанию
	if req.LocalPort == 0 {
		req.LocalPort = 1080
	}
	if req.IATMode == "" {
		req.IATMode = "0"
	}

	// Создаем Obfs4 bypass
	bypassReq := map[string]interface{}{
		"name":       req.Name,
		"method":     "obfs4",
		"local_port": req.LocalPort,
		"bridge":     req.Bridge,
		"cert":       req.Cert,
		"iat_mode":   req.IATMode,
	}

	bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
	if err != nil {
		h.logger.Error("failed to create Obfs4 bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create Obfs4 bypass")
		return
	}

	// Запускаем bypass
	bypassID := getStringFromMap(bypassResp, "id")
	if bypassID == "" {
		writeError(w, http.StatusInternalServerError, "Invalid bypass response")
		return
	}

	if err := h.proxyService.StartBypass(r.Context(), bypassID); err != nil {
		h.logger.Error("failed to start Obfs4 bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start Obfs4 bypass")
		return
	}

	// Формируем ответ
	response := Obfs4ConnectResponse{
		ConnectionID: bypassID,
		LocalPort:    req.LocalPort,
		Status:       "connected",
		Config: Obfs4Config{
			Bridge:    req.Bridge,
			Transport: "obfs4",
			LocalPort: req.LocalPort,
		},
		CreatedAt: time.Now(),
	}

	h.logger.Info("Obfs4 bypass created successfully",
		zap.String("bypass_id", bypassID),
		zap.String("name", req.Name))

	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode Obfs4 connect response")
}

// buildV2RayConfig создает конфигурацию V2Ray
func buildV2RayConfig(req V2RayConnectRequest) V2RayConfig {
	// Простая базовая конфигурация V2Ray
	inbound := map[string]interface{}{
		"port":     req.LocalPort,
		"protocol": "socks",
		"settings": map[string]interface{}{
			"auth": "noauth",
		},
	}

	outbound := map[string]interface{}{
		"protocol": "vmess",
		"settings": map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": req.ServerHost,
					"port":    req.ServerPort,
					"users": []map[string]interface{}{
						{
							"id":       req.UUID,
							"alterId":  req.AlterID,
							"security": req.Security,
						},
					},
				},
			},
		},
		"streamSettings": map[string]interface{}{
			"network": req.Network,
		},
	}

	routing := map[string]interface{}{
		"strategy": "rules",
		"settings": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"outboundTag": "direct",
					"ip":          []string{"geoip:private"},
				},
			},
		},
	}

	return V2RayConfig{
		Inbounds:  []interface{}{inbound},
		Outbounds: []interface{}{outbound},
		Routing:   routing,
	}
}
