package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"go.uber.org/zap"
)

// Handlers HTTP обработчики для Server Manager
type Handlers struct {
	HealthService ports.HealthService
	ServerService ports.ServerService
	Logger        *zap.Logger
}

// NewHandlers создает новые HTTP обработчики
func NewHandlers(health ports.HealthService, server ports.ServerService, logger *zap.Logger) *Handlers {
	return &Handlers{
		HealthService: health,
		ServerService: server,
		Logger:        logger,
	}
}

// RegisterRoutes регистрирует маршруты
func (h *Handlers) RegisterRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", h.HealthHandler).Methods("GET")
	router.HandleFunc("/", h.RootHandler).Methods("GET")

	// API v1
	api := router.PathPrefix("/api/v1").Subrouter()

	// Серверы
	api.HandleFunc("/servers", h.CreateServer).Methods("POST")
	api.HandleFunc("/servers", h.ListServers).Methods("GET")
	api.HandleFunc("/servers/{id}", h.GetServer).Methods("GET")
	api.HandleFunc("/servers/{id}", h.UpdateServer).Methods("PUT")
	api.HandleFunc("/servers/{id}", h.DeleteServer).Methods("DELETE")
	api.HandleFunc("/servers/{id}/start", h.StartServer).Methods("POST")
	api.HandleFunc("/servers/{id}/stop", h.StopServer).Methods("POST")
	api.HandleFunc("/servers/{id}/restart", h.RestartServer).Methods("POST")
	api.HandleFunc("/servers/{id}/stats", h.GetServerStats).Methods("GET")
	api.HandleFunc("/servers/{id}/health", h.GetServerHealth).Methods("GET")

	// Масштабирование
	api.HandleFunc("/scaling/policies", h.GetScalingPolicies).Methods("GET")
	api.HandleFunc("/scaling/policies", h.CreateScalingPolicy).Methods("POST")
	api.HandleFunc("/scaling/policies/{id}", h.UpdateScalingPolicy).Methods("PUT")
	api.HandleFunc("/scaling/policies/{id}", h.DeleteScalingPolicy).Methods("DELETE")
	api.HandleFunc("/scaling/evaluate", h.EvaluateScaling).Methods("POST")

	// Резервное копирование
	api.HandleFunc("/backups/configs", h.GetBackupConfigs).Methods("GET")
	api.HandleFunc("/backups/configs", h.CreateBackupConfig).Methods("POST")
	api.HandleFunc("/backups/configs/{id}", h.UpdateBackupConfig).Methods("PUT")
	api.HandleFunc("/backups/configs/{id}", h.DeleteBackupConfig).Methods("DELETE")
	api.HandleFunc("/servers/{id}/backup", h.CreateBackup).Methods("POST")
	api.HandleFunc("/servers/{id}/restore/{backup_id}", h.RestoreBackup).Methods("POST")

	// Обновления
	api.HandleFunc("/servers/{id}/update", h.GetUpdateStatus).Methods("GET")
	api.HandleFunc("/servers/{id}/update", h.StartUpdate).Methods("POST")
	api.HandleFunc("/servers/{id}/update/cancel", h.CancelUpdate).Methods("POST")

	// Мониторинг
	api.HandleFunc("/health/all", h.GetAllServersHealth).Methods("GET")
}

// HealthHandler обработчик health check
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := h.HealthService.GetHealth()
	h.writeJSON(w, http.StatusOK, health)
}

// RootHandler обработчик корневого маршрута
func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Silence Server Manager Service",
		"version": "1.0.0",
	})
}

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

// GetScalingPolicies получает политики масштабирования
func (h *Handlers) GetScalingPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.ServerService.GetScalingPolicies(r.Context())
	if err != nil {
		h.Logger.Error("failed to get scaling policies", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get scaling policies")
		return
	}

	h.writeJSON(w, http.StatusOK, policies)
}

// CreateScalingPolicy создает политику масштабирования
func (h *Handlers) CreateScalingPolicy(w http.ResponseWriter, r *http.Request) {
	var policy domain.ScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ServerService.CreateScalingPolicy(r.Context(), &policy); err != nil {
		h.Logger.Error("failed to create scaling policy", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create scaling policy")
		return
	}

	h.writeJSON(w, http.StatusCreated, policy)
}

// UpdateScalingPolicy обновляет политику масштабирования
func (h *Handlers) UpdateScalingPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var policy domain.ScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ServerService.UpdateScalingPolicy(r.Context(), id, &policy); err != nil {
		h.Logger.Error("failed to update scaling policy", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to update scaling policy")
		return
	}

	h.writeJSON(w, http.StatusOK, policy)
}

// DeleteScalingPolicy удаляет политику масштабирования
func (h *Handlers) DeleteScalingPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.ServerService.DeleteScalingPolicy(r.Context(), id); err != nil {
		h.Logger.Error("failed to delete scaling policy", zap.String("id", id), zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to delete scaling policy")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "scaling policy deleted"})
}

// EvaluateScaling оценивает необходимость масштабирования
func (h *Handlers) EvaluateScaling(w http.ResponseWriter, r *http.Request) {
	if err := h.ServerService.EvaluateScaling(r.Context()); err != nil {
		h.Logger.Error("failed to evaluate scaling", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to evaluate scaling")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "scaling evaluation completed"})
}

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

// writeJSON записывает JSON ответ
func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.Logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

// writeError записывает ошибку
func (h *Handlers) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
