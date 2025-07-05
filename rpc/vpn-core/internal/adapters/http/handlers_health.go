package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := h.healthService.GetHealth()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
