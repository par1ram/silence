package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

func (h *AnalyticsHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard domain.DashboardConfig
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		h.logger.Error("Failed to decode dashboard", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.analyticsService.CreateDashboard(r.Context(), dashboard); err != nil {
		h.logger.Error("Failed to create dashboard", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dashboard, err := h.analyticsService.GetDashboard(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get dashboard", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, dashboard)
}
func (h *AnalyticsHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var dashboard domain.DashboardConfig
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		h.logger.Error("Failed to decode dashboard", zap.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	dashboard.ID = id
	if err := h.analyticsService.UpdateDashboard(r.Context(), dashboard); err != nil {
		h.logger.Error("Failed to update dashboard", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *AnalyticsHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if err := h.analyticsService.DeleteDashboard(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete dashboard", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *AnalyticsHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	dashboards, err := h.analyticsService.ListDashboards(r.Context())
	if err != nil {
		h.logger.Error("Failed to list dashboards", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, dashboards)
}
