package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// GetBackupConfigs получает конфигурации резервного копирования
func (h *Handlers) GetBackupConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.ServerService.GetBackupConfigs(r.Context())
	if err != nil {
		h.Logger.Error("failed to get backup configs", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get backup configs")
		return
	}

	h.writeJSON(w, http.StatusOK, configs)
}

// CreateBackupConfig создает конфигурацию резервного копирования
func (h *Handlers) CreateBackupConfig(w http.ResponseWriter, r *http.Request) {
	var config domain.BackupConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ServerService.CreateBackupConfig(r.Context(), &config); err != nil {
		h.Logger.Error("failed to create backup config", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create backup config")
		return
	}

	h.writeJSON(w, http.StatusCreated, config)
}

// UpdateBackupConfig обновляет конфигурацию резервного копирования
func (h *Handlers) UpdateBackupConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var config domain.BackupConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ServerService.UpdateBackupConfig(r.Context(), id, &config); err != nil {
		h.Logger.Error("failed to update backup config", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to update backup config")
		return
	}

	h.writeJSON(w, http.StatusOK, config)
}

// DeleteBackupConfig удаляет конфигурацию резервного копирования
func (h *Handlers) DeleteBackupConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.DeleteBackupConfig(r.Context(), id); err != nil {
		h.Logger.Error("failed to delete backup config", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to delete backup config")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "backup config deleted"})
}

// CreateBackup создает резервную копию
func (h *Handlers) CreateBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["id"]

	if err := h.ServerService.CreateBackup(r.Context(), serverID); err != nil {
		h.Logger.Error("failed to create backup", zap.String("server_id", serverID), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create backup")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "backup created"})
}

// RestoreBackup восстанавливает из резервной копии
func (h *Handlers) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["id"]
	backupID := vars["backup_id"]

	if err := h.ServerService.RestoreBackup(r.Context(), serverID, backupID); err != nil {
		h.Logger.Error("failed to restore backup", zap.String("server_id", serverID), zap.String("backup_id", backupID), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to restore backup")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "backup restored"})
}
