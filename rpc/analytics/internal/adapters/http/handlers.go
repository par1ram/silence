package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
)

// AnalyticsHandler HTTP обработчики для аналитики
// В этом файле только структура, конструктор, RegisterRoutes и вспомогательные методы

type AnalyticsHandler struct {
	analyticsService ports.AnalyticsService
	logger           *zap.Logger
}

// NewAnalyticsHandler создает новый обработчик аналитики
func NewAnalyticsHandler(analyticsService ports.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// RegisterRoutes регистрирует маршруты
func (h *AnalyticsHandler) RegisterRoutes(router *mux.Router) {
	// Метрики подключений
	router.HandleFunc("/api/v1/analytics/connections", h.RecordConnection).Methods("POST")
	router.HandleFunc("/api/v1/analytics/connections", h.GetConnectionAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/analytics/connections/stats", h.GetConnectionStats).Methods("GET")

	// Метрики обхода DPI
	router.HandleFunc("/api/v1/analytics/bypass-effectiveness", h.RecordBypassEffectiveness).Methods("POST")
	router.HandleFunc("/api/v1/analytics/bypass-effectiveness", h.GetBypassEffectivenessAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/analytics/bypass-effectiveness/stats", h.GetBypassEffectivenessStats).Methods("GET")

	// Метрики активности пользователей
	router.HandleFunc("/api/v1/analytics/user-activity", h.RecordUserActivity).Methods("POST")
	router.HandleFunc("/api/v1/analytics/user-activity", h.GetUserActivityAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/analytics/user-activity/stats", h.GetUserActivityStats).Methods("GET")

	// Метрики нагрузки серверов
	router.HandleFunc("/api/v1/analytics/server-load", h.RecordServerLoad).Methods("POST")
	router.HandleFunc("/api/v1/analytics/server-load", h.GetServerLoadAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/analytics/server-load/stats", h.GetServerLoadStats).Methods("GET")

	// Метрики ошибок
	router.HandleFunc("/api/v1/analytics/errors", h.RecordError).Methods("POST")
	router.HandleFunc("/api/v1/analytics/errors", h.GetErrorAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/analytics/errors/stats", h.GetErrorStats).Methods("GET")

	// Временные серии
	router.HandleFunc("/api/v1/analytics/timeseries/{metricName}", h.GetTimeSeries).Methods("GET")

	// Дашборды
	router.HandleFunc("/api/v1/analytics/dashboards", h.CreateDashboard).Methods("POST")
	router.HandleFunc("/api/v1/analytics/dashboards", h.ListDashboards).Methods("GET")
	router.HandleFunc("/api/v1/analytics/dashboards/{id}", h.GetDashboard).Methods("GET")
	router.HandleFunc("/api/v1/analytics/dashboards/{id}", h.UpdateDashboard).Methods("PUT")
	router.HandleFunc("/api/v1/analytics/dashboards/{id}", h.DeleteDashboard).Methods("DELETE")

	// Прогнозирование
	router.HandleFunc("/api/v1/analytics/predict/load/{serverID}", h.PredictLoad).Methods("GET")
	router.HandleFunc("/api/v1/analytics/predict/bypass-effectiveness/{bypassType}", h.PredictBypassEffectiveness).Methods("GET")
}

// parseQueryOptions парсит параметры запроса
func (h *AnalyticsHandler) parseQueryOptions(r *http.Request) (domain.QueryOptions, error) {
	opts := domain.QueryOptions{
		Filters: make(map[string]string),
	}

	// Парсим временной диапазон
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			opts.TimeRange.Start = start
		}
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			opts.TimeRange.End = end
		}
	}

	// Парсим агрегацию
	if agg := r.URL.Query().Get("aggregation"); agg != "" {
		opts.Aggregation = domain.AggregationType(agg)
	}

	// Парсим группировку
	if groupBy := r.URL.Query().Get("group_by"); groupBy != "" {
		opts.GroupBy = []string{groupBy}
	}

	// Парсим лимит и смещение
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			opts.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			opts.Offset = offset
		}
	}

	// Парсим интервал для временных серий
	if interval := r.URL.Query().Get("interval"); interval != "" {
		opts.Interval = interval
	}

	// Парсим фильтры
	for key, values := range r.URL.Query() {
		if key != "start" && key != "end" && key != "aggregation" && key != "group_by" && key != "limit" && key != "offset" && key != "interval" {
			if len(values) > 0 {
				opts.Filters[key] = values[0]
			}
		}
	}

	return opts, nil
}

// writeJSON записывает JSON ответ
func (h *AnalyticsHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
