package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// GetUpdateStatus получает статус обновления
func (h *Handlers) GetUpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["id"]

	status, err := h.ServerService.GetUpdateStatus(r.Context(), serverID)
	if err != nil {
		h.Logger.Error("failed to get update status", zap.String("server_id", serverID), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get update status")
		return
	}

	h.writeJSON(w, http.StatusOK, status)
}

// StartUpdate запускает обновление
func (h *Handlers) StartUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["id"]

	var req domain.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ServerID = serverID

	if err := h.ServerService.StartUpdate(r.Context(), &req); err != nil {
		h.Logger.Error("failed to start update", zap.String("server_id", serverID), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to start update")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "update started"})
}

// CancelUpdate отменяет обновление
func (h *Handlers) CancelUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["id"]

	if err := h.ServerService.CancelUpdate(r.Context(), serverID); err != nil {
		h.Logger.Error("failed to cancel update", zap.String("server_id", serverID), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to cancel update")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "update canceled"})
}
