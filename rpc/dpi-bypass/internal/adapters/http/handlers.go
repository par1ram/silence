package http

import (
	"encoding/json"
	"net/http"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/ports"
	"go.uber.org/zap"
)

// Handlers HTTP обработчики для DPI Bypass
// Каждый эндпоинт — отдельная функция

type Handlers struct {
	HealthService ports.HealthService
	BypassService ports.BypassService
	Logger        *zap.Logger
}

func NewHandlers(health ports.HealthService, bypass ports.BypassService, logger *zap.Logger) *Handlers {
	return &Handlers{
		HealthService: health,
		BypassService: bypass,
		Logger:        logger,
	}
}

// Health
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := h.HealthService.GetHealth()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}

// Root
func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Silence DPI Bypass Service"})
}

// POST /api/v1/bypass
func (h *Handlers) CreateBypass(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateBypassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	config, err := h.BypassService.CreateBypass(r.Context(), &req)
	if err != nil {
		http.Error(w, "failed to create bypass", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(config)
}

// GET /api/v1/bypass
func (h *Handlers) ListBypasses(w http.ResponseWriter, r *http.Request) {
	configs, err := h.BypassService.ListBypasses(r.Context())
	if err != nil {
		http.Error(w, "failed to list bypasses", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(configs)
}

// GET /api/v1/bypass/{id}
func (h *Handlers) GetBypass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	config, err := h.BypassService.GetBypass(r.Context(), id)
	if err != nil {
		http.Error(w, "bypass not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

// DELETE /api/v1/bypass/{id}
func (h *Handlers) DeleteBypass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	err := h.BypassService.DeleteBypass(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to delete bypass", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/bypass/{id}/start
func (h *Handlers) StartBypass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	err := h.BypassService.StartBypass(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to start bypass", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"started"}`))
}

// POST /api/v1/bypass/{id}/stop
func (h *Handlers) StopBypass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	err := h.BypassService.StopBypass(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to stop bypass", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"stopped"}`))
}

// GET /api/v1/bypass/{id}/stats
func (h *Handlers) GetBypassStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	stats, err := h.BypassService.GetBypassStats(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
