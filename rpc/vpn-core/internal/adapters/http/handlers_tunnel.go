package http

import (
	"encoding/json"
	"net/http"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"go.uber.org/zap"
)

// CreateTunnelHandler создает новый туннель
func (h *Handlers) CreateTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tunnel, err := h.tunnelManager.CreateTunnel(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create tunnel", zap.Error(err))
		http.Error(w, "Failed to create tunnel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(tunnel); err != nil {
		h.logger.Error("failed to encode tunnel response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetTunnelHandler получает туннель по ID
func (h *Handlers) GetTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	tunnel, err := h.tunnelManager.GetTunnel(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get tunnel", zap.Error(err), zap.String("id", id))
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tunnel); err != nil {
		h.logger.Error("failed to encode tunnel response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ListTunnelsHandler возвращает список всех туннелей
func (h *Handlers) ListTunnelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnels, err := h.tunnelManager.ListTunnels(r.Context())
	if err != nil {
		h.logger.Error("failed to list tunnels", zap.Error(err))
		http.Error(w, "Failed to list tunnels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tunnels); err != nil {
		h.logger.Error("failed to encode tunnels response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// StartTunnelHandler запускает туннель
func (h *Handlers) StartTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	if err := h.tunnelManager.StartTunnel(r.Context(), id); err != nil {
		h.logger.Error("failed to start tunnel", zap.Error(err), zap.String("id", id))
		http.Error(w, "Failed to start tunnel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StopTunnelHandler останавливает туннель
func (h *Handlers) StopTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	if err := h.tunnelManager.StopTunnel(r.Context(), id); err != nil {
		h.logger.Error("failed to stop tunnel", zap.Error(err), zap.String("id", id))
		http.Error(w, "Failed to stop tunnel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetTunnelStatsHandler получает статистику туннеля
func (h *Handlers) GetTunnelStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	stats, err := h.tunnelManager.GetTunnelStats(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get tunnel stats", zap.Error(err), zap.String("id", id))
		http.Error(w, "Failed to get tunnel stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode stats response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
