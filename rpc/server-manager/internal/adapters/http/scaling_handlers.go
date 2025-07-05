package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

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
