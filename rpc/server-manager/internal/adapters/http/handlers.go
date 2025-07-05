package http

import (
	"encoding/json"
	"net/http"
	"strings"

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

// RegisterRoutes регистрирует маршруты через стандартный ServeMux
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc("/", h.RootHandler)

	// API v1
	mux.HandleFunc("/api/v1/servers", h.ServersHandler)
	mux.HandleFunc("/api/v1/servers/", h.ServerByIDHandler)
	mux.HandleFunc("/api/v1/scaling/policies", h.ScalingPoliciesHandler)
	mux.HandleFunc("/api/v1/scaling/policies/", h.ScalingPolicyByIDHandler)
	mux.HandleFunc("/api/v1/scaling/evaluate", h.EvaluateScaling)
	mux.HandleFunc("/api/v1/backups/configs", h.BackupConfigsHandler)
	mux.HandleFunc("/api/v1/backups/configs/", h.BackupConfigByIDHandler)
	mux.HandleFunc("/api/v1/servers/", h.ServerActionsHandler)
	mux.HandleFunc("/api/v1/health/all", h.GetAllServersHealth)
}

// HealthHandler обработчик health check
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	health := h.HealthService.GetHealth()
	h.writeJSON(w, http.StatusOK, health)
}

// RootHandler обработчик корневого маршрута
func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Silence Server Manager Service",
		"version": "1.0.0",
	})
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

// --- Ниже идут новые хендлеры для ServeMux ---

// ServersHandler обрабатывает /api/v1/servers (POST, GET)
func (h *Handlers) ServersHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/servers" {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method == http.MethodPost {
		h.CreateServer(w, r)
		return
	}
	if r.Method == http.MethodGet {
		h.ListServers(w, r)
		return
	}
	h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// ServerByIDHandler обрабатывает /api/v1/servers/{id} и вложенные пути
func (h *Handlers) ServerByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/servers/") {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/servers/")
	parts := strings.Split(path, "/")
	id := parts[0]
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing server id")
		return
	}
	// /api/v1/servers/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetServer(w, r)
		case http.MethodPut:
			h.UpdateServer(w, r)
		case http.MethodDelete:
			h.DeleteServer(w, r)
		default:
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}
	// /api/v1/servers/{id}/...
	if len(parts) >= 2 {
		switch parts[1] {
		case "start":
			if r.Method == http.MethodPost {
				h.StartServer(w, r)
				return
			}
		case "stop":
			if r.Method == http.MethodPost {
				h.StopServer(w, r)
				return
			}
		case "restart":
			if r.Method == http.MethodPost {
				h.RestartServer(w, r)
				return
			}
		case "stats":
			if r.Method == http.MethodGet {
				h.GetServerStats(w, r)
				return
			}
		case "health":
			if r.Method == http.MethodGet {
				h.GetServerHealth(w, r)
				return
			}
		case "backup":
			if r.Method == http.MethodPost {
				h.CreateBackup(w, r)
				return
			}
		case "restore":
			if r.Method == http.MethodPost && len(parts) == 3 {
				h.RestoreBackup(w, r)
				return
			}
		case "update":
			if r.Method == http.MethodGet {
				h.GetUpdateStatus(w, r)
				return
			}
			if r.Method == http.MethodPost {
				h.StartUpdate(w, r)
				return
			}
			if r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "cancel" {
				h.CancelUpdate(w, r)
				return
			}
		}
	}
	h.writeError(w, http.StatusNotFound, "not found")
}

// ScalingPoliciesHandler обрабатывает /api/v1/scaling/policies (GET, POST)
func (h *Handlers) ScalingPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/scaling/policies" {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method == http.MethodGet {
		h.GetScalingPolicies(w, r)
		return
	}
	if r.Method == http.MethodPost {
		h.CreateScalingPolicy(w, r)
		return
	}
	h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// ScalingPolicyByIDHandler обрабатывает /api/v1/scaling/policies/{id}
func (h *Handlers) ScalingPolicyByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/scaling/policies/") {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scaling/policies/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing policy id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		h.UpdateScalingPolicy(w, r)
	case http.MethodDelete:
		h.DeleteScalingPolicy(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// BackupConfigsHandler обрабатывает /api/v1/backups/configs (GET, POST)
func (h *Handlers) BackupConfigsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/backups/configs" {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method == http.MethodGet {
		h.GetBackupConfigs(w, r)
		return
	}
	if r.Method == http.MethodPost {
		h.CreateBackupConfig(w, r)
		return
	}
	h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// BackupConfigByIDHandler обрабатывает /api/v1/backups/configs/{id}
func (h *Handlers) BackupConfigByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/backups/configs/") {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/backups/configs/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing backup config id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		h.UpdateBackupConfig(w, r)
	case http.MethodDelete:
		h.DeleteBackupConfig(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// ServerActionsHandler обрабатывает вложенные действия для серверов (backup, restore, update, ...)
func (h *Handlers) ServerActionsHandler(w http.ResponseWriter, r *http.Request) {
	// Этот хендлер нужен для вложенных путей типа /api/v1/servers/{id}/...
	// Реализация делегирована в ServerByIDHandler
	h.ServerByIDHandler(w, r)
}
