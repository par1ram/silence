package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DisconnectHandler обрабатывает отключение соединений
func (h *Handlers) DisconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var disconnected []string
	var errors []string

	// Обрабатываем разные сценарии отключения
	if req.All {
		// Отключаем все активные соединения
		tunnels, err := h.getActiveTunnels(r.Context())
		if err == nil {
			for _, tunnel := range tunnels {
				if err := h.proxyService.StopVPNTunnel(r.Context(), tunnel.ID); err == nil {
					disconnected = append(disconnected, tunnel.ID)
				} else {
					errors = append(errors, fmt.Sprintf("Failed to stop tunnel %s: %v", tunnel.ID, err))
				}
			}
		}

		bypasses, err := h.getActiveBypasses(r.Context())
		if err == nil {
			for _, bypass := range bypasses {
				if err := h.proxyService.StopBypass(r.Context(), bypass.ID); err == nil {
					disconnected = append(disconnected, bypass.ID)
				} else {
					errors = append(errors, fmt.Sprintf("Failed to stop bypass %s: %v", bypass.ID, err))
				}
			}
		}
	} else {
		// Отключаем конкретные соединения
		if req.ConnectionID != "" {
			if err := h.stopConnection(r.Context(), req.ConnectionID); err == nil {
				disconnected = append(disconnected, req.ConnectionID)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to stop connection %s: %v", req.ConnectionID, err))
			}
		}

		if req.TunnelID != "" {
			if err := h.proxyService.StopVPNTunnel(r.Context(), req.TunnelID); err == nil {
				disconnected = append(disconnected, req.TunnelID)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to stop tunnel %s: %v", req.TunnelID, err))
			}
		}

		if req.BypassID != "" {
			if err := h.proxyService.StopBypass(r.Context(), req.BypassID); err == nil {
				disconnected = append(disconnected, req.BypassID)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to stop bypass %s: %v", req.BypassID, err))
			}
		}
	}

	// Определяем статус ответа
	status := "success"
	if len(errors) > 0 {
		if len(disconnected) > 0 {
			status = "partial"
		} else {
			status = "error"
		}
	}

	response := DisconnectResponse{
		Disconnected: disconnected,
		Status:       status,
		Errors:       errors,
	}

	h.logger.Info("Disconnect operation completed",
		zap.String("status", status),
		zap.Int("disconnected_count", len(disconnected)),
		zap.Int("errors_count", len(errors)))

	writeJSON(w, http.StatusOK, response, h.logger, "failed to encode disconnect response")
}

// ConnectionStatusHandler возвращает статус всех соединений
func (h *Handlers) ConnectionStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем активные VPN туннели
	vpnTunnels := []VPNTunnelStatus{}
	tunnels, err := h.getActiveTunnels(r.Context())
	if err == nil {
		for _, tunnel := range tunnels {
			stats := h.getTunnelStats(r.Context(), tunnel.ID)
			vpnTunnels = append(vpnTunnels, VPNTunnelStatus{
				TunnelID:    tunnel.ID,
				Name:        tunnel.Name,
				Status:      tunnel.Status,
				ConnectedAt: tunnel.CreatedAt,
				PeersCount:  stats.PeersCount,
				BytesRx:     stats.BytesRx,
				BytesTx:     stats.BytesTx,
			})
		}
	}

	// Получаем активные DPI bypasses
	dpiBypasses := []DPIBypassStatus{}
	bypasses, err := h.getActiveBypasses(r.Context())
	if err == nil {
		for _, bypass := range bypasses {
			stats := h.getBypassStats(r.Context(), bypass.ID)
			dpiBypasses = append(dpiBypasses, DPIBypassStatus{
				BypassID:    bypass.ID,
				Name:        bypass.Name,
				Method:      bypass.Method,
				Status:      bypass.Status,
				LocalPort:   bypass.LocalPort,
				ConnectedAt: bypass.CreatedAt,
				BytesRx:     stats.BytesRx,
				BytesTx:     stats.BytesTx,
			})
		}
	}

	// Вычисляем общие показатели
	totalDataTransferred := int64(0)
	for _, tunnel := range vpnTunnels {
		totalDataTransferred += tunnel.BytesRx + tunnel.BytesTx
	}
	for _, bypass := range dpiBypasses {
		totalDataTransferred += bypass.BytesRx + bypass.BytesTx
	}

	response := ConnectionStatusResponse{
		VPNTunnels:           vpnTunnels,
		DPIBypasses:          dpiBypasses,
		ActiveConnections:    len(vpnTunnels) + len(dpiBypasses),
		TotalDataTransferred: totalDataTransferred,
		Uptime:               h.calculateUptime(vpnTunnels, dpiBypasses),
	}

	writeJSON(w, http.StatusOK, response, h.logger, "failed to encode connection status response")
}

// Вспомогательные методы для получения данных

type TunnelInfo struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
}

type BypassInfo struct {
	ID        string
	Name      string
	Method    string
	Status    string
	LocalPort int
	CreatedAt time.Time
}

type TunnelStats struct {
	PeersCount int
	BytesRx    int64
	BytesTx    int64
}

type BypassStats struct {
	BytesRx int64
	BytesTx int64
}

func (h *Handlers) getActiveTunnels(ctx interface{}) ([]TunnelInfo, error) {
	// Здесь должен быть вызов к VPN Core сервису
	// Пока возвращаем пустой список
	return []TunnelInfo{}, nil
}

func (h *Handlers) getActiveBypasses(ctx interface{}) ([]BypassInfo, error) {
	// Здесь должен быть вызов к DPI Bypass сервису
	// Пока возвращаем пустой список
	return []BypassInfo{}, nil
}

func (h *Handlers) getTunnelStats(ctx interface{}, tunnelID string) TunnelStats {
	// Здесь должен быть вызов к VPN Core сервису для получения статистики
	return TunnelStats{}
}

func (h *Handlers) getBypassStats(ctx interface{}, bypassID string) BypassStats {
	// Здесь должен быть вызов к DPI Bypass сервису для получения статистики
	return BypassStats{}
}

func (h *Handlers) stopConnection(ctx interface{}, connectionID string) error {
	// Здесь должна быть логика для определения типа соединения и его остановки
	return nil
}

func (h *Handlers) calculateUptime(vpnTunnels []VPNTunnelStatus, dpiBypasses []DPIBypassStatus) int64 {
	if len(vpnTunnels) == 0 && len(dpiBypasses) == 0 {
		return 0
	}

	var earliestTime time.Time
	first := true

	for _, tunnel := range vpnTunnels {
		if first || tunnel.ConnectedAt.Before(earliestTime) {
			earliestTime = tunnel.ConnectedAt
			first = false
		}
	}

	for _, bypass := range dpiBypasses {
		if first || bypass.ConnectedAt.Before(earliestTime) {
			earliestTime = bypass.ConnectedAt
			first = false
		}
	}

	if first {
		return 0
	}

	return int64(time.Since(earliestTime).Seconds())
}
