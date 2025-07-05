package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

// Метрики подключений
func (h *AnalyticsHandler) RecordConnection(w http.ResponseWriter, r *http.Request) {
	var metric domain.ConnectionMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Error("Failed to decode connection metric", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.RecordConnection(r.Context(), metric); err != nil {
		h.logger.Error("Failed to record connection metric", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetConnectionAnalytics(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	response, err := h.analyticsService.GetConnectionAnalytics(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get connection analytics", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, response)
}
func (h *AnalyticsHandler) GetConnectionStats(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	stats, err := h.analyticsService.GetConnectionStats(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get connection stats", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, stats)
}

// Метрики обхода DPI
func (h *AnalyticsHandler) RecordBypassEffectiveness(w http.ResponseWriter, r *http.Request) {
	var metric domain.BypassEffectivenessMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Error("Failed to decode bypass effectiveness metric", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.RecordBypassEffectiveness(r.Context(), metric); err != nil {
		h.logger.Error("Failed to record bypass effectiveness metric", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetBypassEffectivenessAnalytics(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	response, err := h.analyticsService.GetBypassEffectivenessAnalytics(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get bypass effectiveness analytics", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, response)
}
func (h *AnalyticsHandler) GetBypassEffectivenessStats(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	stats, err := h.analyticsService.GetBypassEffectivenessStats(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get bypass effectiveness stats", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, stats)
}

// Метрики активности пользователей
func (h *AnalyticsHandler) RecordUserActivity(w http.ResponseWriter, r *http.Request) {
	var metric domain.UserActivityMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Error("Failed to decode user activity metric", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.RecordUserActivity(r.Context(), metric); err != nil {
		h.logger.Error("Failed to record user activity metric", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetUserActivityAnalytics(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	response, err := h.analyticsService.GetUserActivityAnalytics(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get user activity analytics", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, response)
}
func (h *AnalyticsHandler) GetUserActivityStats(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	stats, err := h.analyticsService.GetUserActivityStats(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get user activity stats", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, stats)
}

// Метрики нагрузки серверов
func (h *AnalyticsHandler) RecordServerLoad(w http.ResponseWriter, r *http.Request) {
	var metric domain.ServerLoadMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Error("Failed to decode server load metric", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.RecordServerLoad(r.Context(), metric); err != nil {
		h.logger.Error("Failed to record server load metric", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetServerLoadAnalytics(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	response, err := h.analyticsService.GetServerLoadAnalytics(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get server load analytics", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, response)
}
func (h *AnalyticsHandler) GetServerLoadStats(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	stats, err := h.analyticsService.GetServerLoadStats(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get server load stats", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, stats)
}

// Метрики ошибок
func (h *AnalyticsHandler) RecordError(w http.ResponseWriter, r *http.Request) {
	var metric domain.ErrorMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Error("Failed to decode error metric", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.RecordError(r.Context(), metric); err != nil {
		h.logger.Error("Failed to record error metric", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetErrorAnalytics(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	response, err := h.analyticsService.GetErrorAnalytics(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get error analytics", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, response)
}
func (h *AnalyticsHandler) GetErrorStats(w http.ResponseWriter, r *http.Request) {
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	stats, err := h.analyticsService.GetErrorStats(r.Context(), opts)
	if err != nil {
		h.logger.Error("Failed to get error stats", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, stats)
}

// Временные серии
func (h *AnalyticsHandler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricName := vars["metricName"]
	opts, err := h.parseQueryOptions(r)
	if err != nil {
		h.logger.Error("Failed to parse query options", zap.String("error", err.Error()))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}
	metrics, err := h.analyticsService.GetTimeSeries(r.Context(), metricName, opts)
	if err != nil {
		h.logger.Error("Failed to get time series", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, metrics)
}
