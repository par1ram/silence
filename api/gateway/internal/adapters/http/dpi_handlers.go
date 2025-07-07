package http

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ConnectDPIHandler обрабатывает DPI-only подключения
func (h *Handlers) ConnectDPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DPIConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация запроса
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.RemoteHost == "" {
		writeError(w, http.StatusBadRequest, "Remote host is required")
		return
	}
	if req.RemotePort <= 0 {
		writeError(w, http.StatusBadRequest, "Valid remote port is required")
		return
	}

	// Устанавливаем значения по умолчанию
	if req.LocalPort == 0 {
		req.LocalPort = 1080
	}
	if req.Encryption == "" {
		req.Encryption = "aes-256-gcm"
	}

	// Создаем DPI bypass
	bypassReq := map[string]interface{}{
		"name":        req.Name,
		"method":      req.Method,
		"local_port":  req.LocalPort,
		"remote_host": req.RemoteHost,
		"remote_port": req.RemotePort,
		"password":    req.Password,
		"encryption":  req.Encryption,
	}

	bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
	if err != nil {
		h.logger.Error("failed to create DPI bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create DPI bypass")
		return
	}

	// Запускаем bypass
	bypassID := getStringFromMap(bypassResp, "id")
	if bypassID == "" {
		writeError(w, http.StatusInternalServerError, "Invalid bypass response")
		return
	}

	if err := h.proxyService.StartBypass(r.Context(), bypassID); err != nil {
		h.logger.Error("failed to start DPI bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start DPI bypass")
		return
	}

	// Формируем ответ
	response := DPIConnectResponse{
		BypassID:  bypassID,
		Method:    req.Method,
		LocalPort: req.LocalPort,
		Status:    "running",
		Config:    buildDPIConfig(bypassResp),
		CreatedAt: time.Now(),
	}

	h.logger.Info("DPI bypass created successfully",
		zap.String("bypass_id", bypassID),
		zap.String("name", req.Name),
		zap.String("method", req.Method))

	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode DPI connect response")
}

// ConnectShadowsocksHandler обрабатывает Shadowsocks подключения
func (h *Handlers) ConnectShadowsocksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ShadowsocksConnectRequest
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
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Устанавливаем значения по умолчанию
	if req.LocalPort == 0 {
		req.LocalPort = 1080
	}
	if req.Encryption == "" {
		req.Encryption = "aes-256-gcm"
	}
	if req.Timeout == 0 {
		req.Timeout = 300
	}

	// Создаем Shadowsocks bypass
	bypassReq := map[string]interface{}{
		"name":        req.Name,
		"method":      "shadowsocks",
		"local_port":  req.LocalPort,
		"remote_host": req.ServerHost,
		"remote_port": req.ServerPort,
		"password":    req.Password,
		"encryption":  req.Encryption,
		"timeout":     req.Timeout,
	}

	bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
	if err != nil {
		h.logger.Error("failed to create Shadowsocks bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to create Shadowsocks bypass")
		return
	}

	// Запускаем bypass
	bypassID := getStringFromMap(bypassResp, "id")
	if bypassID == "" {
		writeError(w, http.StatusInternalServerError, "Invalid bypass response")
		return
	}

	if err := h.proxyService.StartBypass(r.Context(), bypassID); err != nil {
		h.logger.Error("failed to start Shadowsocks bypass", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "Failed to start Shadowsocks bypass")
		return
	}

	// Формируем ответ
	response := ShadowsocksConnectResponse{
		ConnectionID: bypassID,
		LocalPort:    req.LocalPort,
		Status:       "connected",
		Config: ShadowsocksConfig{
			Server:     req.ServerHost,
			ServerPort: req.ServerPort,
			LocalPort:  req.LocalPort,
			Password:   req.Password,
			Method:     req.Encryption,
		},
		CreatedAt: time.Now(),
	}

	h.logger.Info("Shadowsocks bypass created successfully",
		zap.String("bypass_id", bypassID),
		zap.String("name", req.Name))

	writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode Shadowsocks connect response")
}

// buildDPIConfig создает конфигурацию DPI из ответа сервиса
func buildDPIConfig(bypassResp map[string]interface{}) map[string]interface{} {
	config := make(map[string]interface{})

	// Копируем основные поля конфигурации
	if method := getStringFromMap(bypassResp, "method"); method != "" {
		config["method"] = method
	}
	if localPort := getIntFromMap(bypassResp, "local_port"); localPort > 0 {
		config["local_port"] = localPort
	}
	if remoteHost := getStringFromMap(bypassResp, "remote_host"); remoteHost != "" {
		config["remote_host"] = remoteHost
	}
	if remotePort := getIntFromMap(bypassResp, "remote_port"); remotePort > 0 {
		config["remote_port"] = remotePort
	}
	if encryption := getStringFromMap(bypassResp, "encryption"); encryption != "" {
		config["encryption"] = encryption
	}

	return config
}

// getIntFromMap извлекает int значение из map
func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return 0
}
