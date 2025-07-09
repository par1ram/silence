package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

func (h *AnalyticsHandler) PredictLoad(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["serverID"]
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // По умолчанию 24 часа
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil {
			hours = h
		}
	}
	predictionReq := &domain.PredictionRequest{
		ServerID:   serverID,
		HoursAhead: hours,
	}
	predictions, err := h.analyticsService.PredictLoad(r.Context(), predictionReq)
	if err != nil {
		h.logger.Error("Failed to predict load", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, predictions)
}

func (h *AnalyticsHandler) PredictBypassEffectiveness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bypassType := vars["bypassType"]
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // По умолчанию 24 часа
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil {
			hours = h
		}
	}
	metrics, err := h.analyticsService.PredictBypassEffectiveness(r.Context(), bypassType, hours)
	if err != nil {
		h.logger.Error("Failed to predict bypass effectiveness", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, metrics)
}
