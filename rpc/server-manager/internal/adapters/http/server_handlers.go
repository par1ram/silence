package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// CreateServer создает новый сервер
func (h *Handlers) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	server, err := h.ServerService.CreateServer(r.Context(), &req)
	if err != nil {
		h.Logger.Error("failed to create server", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create server")
		return
	}

	h.writeJSON(w, http.StatusCreated, server)
}

// GetServer получает сервер по ID
func (h *Handlers) GetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	server, err := h.ServerService.GetServer(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusNotFound, "server not found")
		return
	}

	h.writeJSON(w, http.StatusOK, server)
}

// ListServers получает список серверов
func (h *Handlers) ListServers(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]interface{})

	// Парсим query параметры
	if serverType := r.URL.Query().Get("type"); serverType != "" {
		filters["type"] = domain.ServerType(serverType)
	}
	if region := r.URL.Query().Get("region"); region != "" {
		filters["region"] = region
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = domain.ServerStatus(status)
	}

	servers, err := h.ServerService.ListServers(r.Context(), filters)
	if err != nil {
		h.Logger.Error("failed to list servers", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to list servers")
		return
	}

	h.writeJSON(w, http.StatusOK, servers)
}

// UpdateServer обновляет сервер
func (h *Handlers) UpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	server, err := h.ServerService.UpdateServer(r.Context(), id, &req)
	if err != nil {
		h.Logger.Error("failed to update server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to update server")
		return
	}

	h.writeJSON(w, http.StatusOK, server)
}

// DeleteServer удаляет сервер
func (h *Handlers) DeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.DeleteServer(r.Context(), id); err != nil {
		h.Logger.Error("failed to delete server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to delete server")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "server deleted"})
}

// StartServer запускает сервер
func (h *Handlers) StartServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.StartServer(r.Context(), id); err != nil {
		h.Logger.Error("failed to start server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to start server")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "server started"})
}

// StopServer останавливает сервер
func (h *Handlers) StopServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.StopServer(r.Context(), id); err != nil {
		h.Logger.Error("failed to stop server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to stop server")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "server stopped"})
}

// RestartServer перезапускает сервер
func (h *Handlers) RestartServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.RestartServer(r.Context(), id); err != nil {
		h.Logger.Error("failed to restart server", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to restart server")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "server restarted"})
}

// GetServerStats получает статистику сервера
func (h *Handlers) GetServerStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stats, err := h.ServerService.GetServerStats(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get server stats", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get server stats")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// GetServerHealth получает здоровье сервера
func (h *Handlers) GetServerHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	health, err := h.ServerService.GetServerHealth(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get server health", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get server health")
		return
	}

	h.writeJSON(w, http.StatusOK, health)
}

// GetAllServersHealth получает здоровье всех серверов
func (h *Handlers) GetAllServersHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.ServerService.GetAllServersHealth(r.Context())
	if err != nil {
		h.Logger.Error("failed to get all servers health", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get all servers health")
		return
	}

	h.writeJSON(w, http.StatusOK, health)
}
