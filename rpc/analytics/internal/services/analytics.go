package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"github.com/par1ram/silence/rpc/analytics/internal/telemetry"
	"go.uber.org/zap"
)

// AnalyticsServiceImpl реализация сервиса аналитики с OpenTelemetry
type AnalyticsServiceImpl struct {
	metricsRepo      ports.MetricsRepository
	dashboardRepo    ports.DashboardRepository
	collector        ports.MetricsCollector
	alertService     ports.AlertService
	logger           *zap.Logger
	metricsCollector *telemetry.MetricsCollector
	tracingManager   *telemetry.TracingManager
}

// NewAnalyticsService создает новый сервис аналитики
func NewAnalyticsService(
	metricsRepo ports.MetricsRepository,
	dashboardRepo ports.DashboardRepository,
	collector ports.MetricsCollector,
	alertService ports.AlertService,
	logger *zap.Logger,
	metricsCollector *telemetry.MetricsCollector,
	tracingManager *telemetry.TracingManager,
) ports.AnalyticsService {
	return &AnalyticsServiceImpl{
		metricsRepo:      metricsRepo,
		dashboardRepo:    dashboardRepo,
		collector:        collector,
		alertService:     alertService,
		logger:           logger,
		metricsCollector: metricsCollector,
		tracingManager:   tracingManager,
	}
}

// GetServerLoadMetrics получает метрики нагрузки серверов за период
func (s *AnalyticsServiceImpl) GetServerLoadMetrics(ctx context.Context, start, end time.Time) ([]domain.ServerLoadMetric, error) {
	return s.executeWithTracing(ctx, "GetServerLoadMetrics", func(ctx context.Context) ([]domain.ServerLoadMetric, error) {
		s.tracingManager.SetSpanAttribute(ctx, "time_range.start", start.Format(time.RFC3339))
		s.tracingManager.SetSpanAttribute(ctx, "time_range.end", end.Format(time.RFC3339))
		s.tracingManager.SetSpanAttribute(ctx, "time_range.duration", end.Sub(start).String())

		opts := domain.QueryOptions{
			TimeRange: domain.TimeRange{
				Start: start,
				End:   end,
			},
		}
		metricsResponse, err := s.metricsRepo.GetServerLoadMetrics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "repository_error", "server_load")
			return nil, fmt.Errorf("failed to get server load metrics: %w", err)
		}

		if metricsResponse == nil || metricsResponse.Metrics == nil {
			return []domain.ServerLoadMetric{}, nil
		}

		// Convert MetricResponse to []ServerLoadMetric
		var metrics []domain.ServerLoadMetric
		for _, data := range metricsResponse.Metrics {
			// Convert base Metric to ServerLoadMetric
			serverMetric := domain.ServerLoadMetric{
				Metric: data,
			}
			metrics = append(metrics, serverMetric)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "server_load")
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(metrics))

		return metrics, nil
	})
}

// CollectMetric собирает метрику
func (s *AnalyticsServiceImpl) CollectMetric(ctx context.Context, metric *domain.Metric) (*domain.Metric, error) {
	return s.executeWithTracingMetric(ctx, "CollectMetric", func(ctx context.Context) (*domain.Metric, error) {
		s.tracingManager.SetSpanAttribute(ctx, "metric.name", metric.Name)
		s.tracingManager.SetSpanAttribute(ctx, "metric.type", metric.Type)
		s.tracingManager.SetSpanAttribute(ctx, "metric.value", metric.Value)

		err := s.metricsRepo.SaveMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "save_error", metric.Type)
			return nil, fmt.Errorf("failed to save metric: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, metric.Type)
		s.logger.Info("Metric collected",
			zap.String("name", metric.Name),
			zap.Float64("value", metric.Value),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return metric, nil
	})
}

// GetMetrics получает метрики по фильтрам
func (s *AnalyticsServiceImpl) GetMetrics(ctx context.Context, filters *domain.MetricFilters) ([]*domain.Metric, int, error) {
	return s.executeWithTracingMetrics(ctx, "GetMetrics", func(ctx context.Context) ([]*domain.Metric, int, error) {
		s.tracingManager.SetSpanAttribute(ctx, "filters.name", filters.Name)
		s.tracingManager.SetSpanAttribute(ctx, "filters.limit", filters.Limit)
		s.tracingManager.SetSpanAttribute(ctx, "filters.offset", filters.Offset)

		metrics, count, err := s.metricsRepo.GetMetrics(ctx, filters)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "query_error", "get_metrics")
			return nil, 0, fmt.Errorf("failed to get metrics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "get_metrics")
		s.tracingManager.SetSpanAttribute(ctx, "result.total", count)

		return metrics, count, nil
	})
}

// GetMetricsHistory получает историю метрик
func (s *AnalyticsServiceImpl) GetMetricsHistory(ctx context.Context, req *domain.MetricHistoryRequest) ([]domain.TimeSeriesPoint, error) {
	return s.executeWithTracingTimeSeries(ctx, "GetMetricsHistory", func(ctx context.Context) ([]domain.TimeSeriesPoint, error) {
		s.tracingManager.SetSpanAttribute(ctx, "history.name", req.Name)
		s.tracingManager.SetSpanAttribute(ctx, "history.interval", req.Interval)
		s.tracingManager.SetSpanAttribute(ctx, "history.start", req.StartTime.Format(time.RFC3339))
		s.tracingManager.SetSpanAttribute(ctx, "history.end", req.EndTime.Format(time.RFC3339))

		history, err := s.metricsRepo.GetMetricsHistory(ctx, req)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "history_error", "metrics_history")
			return nil, fmt.Errorf("failed to get metrics history: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "metrics_history")
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(history))

		return history, nil
	})
}

// GetStatistics получает статистику
func (s *AnalyticsServiceImpl) GetStatistics(ctx context.Context, req *domain.StatisticsRequest) ([]*domain.Statistics, error) {
	return s.executeWithTracingStatistics(ctx, "GetStatistics", func(ctx context.Context) ([]*domain.Statistics, error) {
		s.tracingManager.SetSpanAttribute(ctx, "statistics.type", req.Type)
		s.tracingManager.SetSpanAttribute(ctx, "statistics.period", req.Period)

		stats, err := s.metricsRepo.GetStatistics(ctx, req)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "statistics_error", "get_statistics")
			return nil, fmt.Errorf("failed to get statistics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "statistics")
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(stats))

		return stats, nil
	})
}

// GetSystemStats получает системную статистику
func (s *AnalyticsServiceImpl) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	return s.executeWithTracingSystemStats(ctx, "GetSystemStats", func(ctx context.Context) (*domain.SystemStats, error) {
		stats, err := s.metricsRepo.GetSystemStats(ctx)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "system_stats_error", "system_stats")
			return nil, fmt.Errorf("failed to get system stats: %w", err)
		}

		// Записываем системные метрики в OpenTelemetry
		s.metricsCollector.RecordActiveUsers(ctx, stats.ActiveUsers, "global")
		s.metricsCollector.RecordSystemLoad(ctx, stats.SystemLoad)
		s.metricsCollector.RecordMetricProcessed(ctx, "system_stats")

		s.tracingManager.SetSpanAttribute(ctx, "system.active_users", stats.ActiveUsers)
		s.tracingManager.SetSpanAttribute(ctx, "system.active_connections", stats.ActiveConnections)
		s.tracingManager.SetSpanAttribute(ctx, "system.load", stats.SystemLoad)

		return stats, nil
	})
}

// GetUserStats получает статистику пользователя
func (s *AnalyticsServiceImpl) GetUserStats(ctx context.Context, req *domain.UserStatsRequest) (*domain.UserStats, error) {
	return s.executeWithTracingUserStats(ctx, "GetUserStats", func(ctx context.Context) (*domain.UserStats, error) {
		s.tracingManager.SetSpanAttribute(ctx, "user.id", req.UserID)

		stats, err := s.metricsRepo.GetUserStats(ctx, req)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "user_stats_error", "user_stats")
			return nil, fmt.Errorf("failed to get user stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "user_stats")
		s.tracingManager.SetSpanAttribute(ctx, "user.total_connections", stats.TotalConnections)
		s.tracingManager.SetSpanAttribute(ctx, "user.data_transferred", stats.TotalDataTransferred)

		return stats, nil
	})
}

// GetDashboardData получает данные для дашборда
func (s *AnalyticsServiceImpl) GetDashboardData(ctx context.Context, timeRange string) (*domain.DashboardData, error) {
	return s.executeWithTracingDashboard(ctx, "GetDashboardData", func(ctx context.Context) (*domain.DashboardData, error) {
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.time_range", timeRange)
		s.metricsCollector.RecordDashboardRequest(ctx, "main")

		data, err := s.dashboardRepo.GetDashboardData(ctx, timeRange)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "dashboard_error", "dashboard_data")
			return nil, fmt.Errorf("failed to get dashboard data: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "dashboard_data")
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.connections_count", len(data.ConnectionsOverTime))
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.servers_count", len(data.ServerUsage))
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.alerts_count", len(data.Alerts))

		return data, nil
	})
}

// PredictLoad предсказывает нагрузку
func (s *AnalyticsServiceImpl) PredictLoad(ctx context.Context, req *domain.PredictionRequest) ([]domain.PredictionPoint, error) {
	return s.executeWithTracingPrediction(ctx, "PredictLoad", func(ctx context.Context) ([]domain.PredictionPoint, error) {
		s.tracingManager.SetSpanAttribute(ctx, "prediction.server_id", req.ServerID)
		s.tracingManager.SetSpanAttribute(ctx, "prediction.hours_ahead", req.HoursAhead)

		predictions, err := s.metricsRepo.PredictLoad(ctx, req)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "prediction_error", "load_prediction")
			return nil, fmt.Errorf("failed to predict load: %w", err)
		}

		s.metricsCollector.RecordPredictionGenerated(ctx, "load")
		s.tracingManager.SetSpanAttribute(ctx, "prediction.points", len(predictions))

		return predictions, nil
	})
}

// PredictTrend предсказывает тренд
func (s *AnalyticsServiceImpl) PredictTrend(ctx context.Context, req *domain.TrendRequest) ([]domain.PredictionPoint, error) {
	return s.executeWithTracingPrediction(ctx, "PredictTrend", func(ctx context.Context) ([]domain.PredictionPoint, error) {
		s.tracingManager.SetSpanAttribute(ctx, "trend.metric_name", req.MetricName)
		s.tracingManager.SetSpanAttribute(ctx, "trend.days_ahead", req.DaysAhead)

		predictions, err := s.metricsRepo.PredictTrend(ctx, req)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "prediction_error", "trend_prediction")
			return nil, fmt.Errorf("failed to predict trend: %w", err)
		}

		s.metricsCollector.RecordPredictionGenerated(ctx, "trend")
		s.tracingManager.SetSpanAttribute(ctx, "prediction.points", len(predictions))

		return predictions, nil
	})
}

// RecordConnection записывает метрику подключения
func (s *AnalyticsServiceImpl) RecordConnection(ctx context.Context, metric domain.ConnectionMetric) error {
	return s.tracingManager.TraceConnectionMetrics(ctx, metric.UserID, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "connection.user_id", metric.UserID)
		s.tracingManager.SetSpanAttribute(ctx, "connection.server_id", metric.ServerID)
		s.tracingManager.SetSpanAttribute(ctx, "connection.protocol", metric.Protocol)
		s.tracingManager.SetSpanAttribute(ctx, "connection.region", metric.Region)

		err := s.metricsRepo.RecordConnectionMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "record_error", "connection")
			return fmt.Errorf("failed to record connection metric: %w", err)
		}

		// Записываем метрики в OpenTelemetry
		s.metricsCollector.RecordNewConnection(ctx, metric.UserID, metric.ServerID, metric.Protocol, metric.Region)
		s.metricsCollector.RecordConnectionData(ctx, metric.BytesIn, metric.BytesOut, metric.ServerID, metric.Region)
		s.metricsCollector.RecordConnectionDuration(ctx, time.Duration(metric.Duration)*time.Millisecond, metric.ServerID, metric.Region)

		s.logger.Info("Connection metric recorded",
			zap.String("user_id", metric.UserID),
			zap.String("server_id", metric.ServerID),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetConnectionAnalytics получает аналитику подключений
func (s *AnalyticsServiceImpl) GetConnectionAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return s.executeWithTracingMetricResponse(ctx, "GetConnectionAnalytics", func(ctx context.Context) (*domain.MetricResponse, error) {
		s.tracingManager.SetSpanAttribute(ctx, "analytics.type", "connection")
		s.tracingManager.SetSpanAttribute(ctx, "analytics.aggregation", string(opts.Aggregation))

		analytics, err := s.metricsRepo.GetConnectionAnalytics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "analytics_error", "connection_analytics")
			return nil, fmt.Errorf("failed to get connection analytics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "connection_analytics")
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(analytics.Metrics))

		return analytics, nil
	})
}

// GetConnectionStats получает статистику подключений
func (s *AnalyticsServiceImpl) GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return s.executeWithTracingMapInterface(ctx, "GetConnectionStats", func(ctx context.Context) (map[string]interface{}, error) {
		stats, err := s.metricsRepo.GetConnectionStats(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "stats_error", "connection_stats")
			return nil, fmt.Errorf("failed to get connection stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "connection_stats")
		return stats, nil
	})
}

// RecordBypassEffectiveness записывает метрику эффективности обхода
func (s *AnalyticsServiceImpl) RecordBypassEffectiveness(ctx context.Context, metric domain.BypassEffectivenessMetric) error {
	return s.tracingManager.TraceBypassEffectiveness(ctx, metric.BypassType, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "bypass.type", metric.BypassType)
		s.tracingManager.SetSpanAttribute(ctx, "bypass.success_rate", metric.SuccessRate)
		s.tracingManager.SetSpanAttribute(ctx, "bypass.latency", metric.Latency)

		err := s.metricsRepo.RecordBypassEffectivenessMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "record_error", "bypass_effectiveness")
			return fmt.Errorf("failed to record bypass effectiveness metric: %w", err)
		}

		// Записываем метрики в OpenTelemetry
		s.metricsCollector.RecordBypassAttempt(ctx, metric.BypassType, "global")
		if metric.SuccessRate > 0.5 {
			s.metricsCollector.RecordBypassSuccess(ctx, metric.BypassType, "global")
		}
		s.metricsCollector.RecordBypassLatency(ctx, time.Duration(metric.Latency)*time.Millisecond, metric.BypassType, "global")
		s.metricsCollector.RecordBypassThroughput(ctx, metric.Throughput, metric.BypassType, "global")

		s.logger.Info("Bypass effectiveness metric recorded",
			zap.String("bypass_type", metric.BypassType),
			zap.Float64("success_rate", metric.SuccessRate),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetBypassEffectivenessAnalytics получает аналитику эффективности обхода
func (s *AnalyticsServiceImpl) GetBypassEffectivenessAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return s.executeWithTracingMetricResponse(ctx, "GetBypassEffectivenessAnalytics", func(ctx context.Context) (*domain.MetricResponse, error) {
		analytics, err := s.metricsRepo.GetBypassEffectivenessAnalytics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "analytics_error", "bypass_effectiveness")
			return nil, fmt.Errorf("failed to get bypass effectiveness analytics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "bypass_effectiveness_analytics")
		return analytics, nil
	})
}

// GetBypassEffectivenessStats получает статистику эффективности обхода
func (s *AnalyticsServiceImpl) GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return s.executeWithTracingMapInterface(ctx, "GetBypassEffectivenessStats", func(ctx context.Context) (map[string]interface{}, error) {
		stats, err := s.metricsRepo.GetBypassEffectivenessStats(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "stats_error", "bypass_effectiveness_stats")
			return nil, fmt.Errorf("failed to get bypass effectiveness stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "bypass_effectiveness_stats")
		return stats, nil
	})
}

// RecordUserActivity записывает метрику активности пользователя
func (s *AnalyticsServiceImpl) RecordUserActivity(ctx context.Context, metric domain.UserActivityMetric) error {
	return s.tracingManager.TraceUserActivity(ctx, metric.UserID, "activity", func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "user.id", metric.UserID)
		s.tracingManager.SetSpanAttribute(ctx, "user.session_count", metric.SessionCount)
		s.tracingManager.SetSpanAttribute(ctx, "user.total_time", metric.TotalTime)

		err := s.metricsRepo.RecordUserActivityMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "record_error", "user_activity")
			return fmt.Errorf("failed to record user activity metric: %w", err)
		}

		// Записываем метрики в OpenTelemetry
		s.metricsCollector.RecordUserSession(ctx, metric.UserID, "global")
		s.metricsCollector.RecordUserActivityTime(ctx, time.Duration(metric.TotalTime)*time.Minute, metric.UserID, "global")
		s.metricsCollector.RecordUserDataTransferred(ctx, metric.DataUsage*1024*1024, metric.UserID, "global")

		s.logger.Info("User activity metric recorded",
			zap.String("user_id", metric.UserID),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetUserActivityAnalytics получает аналитику активности пользователей
func (s *AnalyticsServiceImpl) GetUserActivityAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return s.executeWithTracingMetricResponse(ctx, "GetUserActivityAnalytics", func(ctx context.Context) (*domain.MetricResponse, error) {
		analytics, err := s.metricsRepo.GetUserActivityAnalytics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "analytics_error", "user_activity")
			return nil, fmt.Errorf("failed to get user activity analytics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "user_activity_analytics")
		return analytics, nil
	})
}

// GetUserActivityStats получает статистику активности пользователей
func (s *AnalyticsServiceImpl) GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return s.executeWithTracingMapInterface(ctx, "GetUserActivityStats", func(ctx context.Context) (map[string]interface{}, error) {
		stats, err := s.metricsRepo.GetUserActivityStats(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "stats_error", "user_activity_stats")
			return nil, fmt.Errorf("failed to get user activity stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "user_activity_stats")
		return stats, nil
	})
}

// RecordServerLoad записывает метрику нагрузки сервера
func (s *AnalyticsServiceImpl) RecordServerLoad(ctx context.Context, metric domain.ServerLoadMetric) error {
	return s.tracingManager.TraceServerLoad(ctx, metric.ServerID, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "server.id", metric.ServerID)
		s.tracingManager.SetSpanAttribute(ctx, "server.region", metric.Region)
		s.tracingManager.SetSpanAttribute(ctx, "server.cpu_usage", metric.CPUUsage)
		s.tracingManager.SetSpanAttribute(ctx, "server.memory_usage", metric.MemoryUsage)

		err := s.metricsRepo.RecordServerLoadMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "record_error", "server_load")
			return fmt.Errorf("failed to record server load metric: %w", err)
		}

		// Записываем метрики в OpenTelemetry
		s.metricsCollector.RecordServerCPUUsage(ctx, metric.CPUUsage, metric.ServerID, metric.Region)
		s.metricsCollector.RecordServerMemoryUsage(ctx, metric.MemoryUsage, metric.ServerID, metric.Region)
		s.metricsCollector.RecordServerNetwork(ctx, metric.NetworkIn, metric.NetworkOut, metric.ServerID, metric.Region)
		s.metricsCollector.RecordServerConnections(ctx, metric.Connections, metric.ServerID, metric.Region)

		s.logger.Info("Server load metric recorded",
			zap.String("server_id", metric.ServerID),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetServerLoadAnalytics получает аналитику нагрузки серверов
func (s *AnalyticsServiceImpl) GetServerLoadAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return s.executeWithTracingMetricResponse(ctx, "GetServerLoadAnalytics", func(ctx context.Context) (*domain.MetricResponse, error) {
		analytics, err := s.metricsRepo.GetServerLoadAnalytics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "analytics_error", "server_load")
			return nil, fmt.Errorf("failed to get server load analytics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "server_load_analytics")
		return analytics, nil
	})
}

// GetServerLoadStats получает статистику нагрузки серверов
func (s *AnalyticsServiceImpl) GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return s.executeWithTracingMapInterface(ctx, "GetServerLoadStats", func(ctx context.Context) (map[string]interface{}, error) {
		stats, err := s.metricsRepo.GetServerLoadStats(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "stats_error", "server_load_stats")
			return nil, fmt.Errorf("failed to get server load stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "server_load_stats")
		return stats, nil
	})
}

// RecordError записывает метрику ошибки
func (s *AnalyticsServiceImpl) RecordError(ctx context.Context, metric domain.ErrorMetric) error {
	return s.tracingManager.TraceAnalyticsOperation(ctx, "RecordError", func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "error.type", metric.ErrorType)
		s.tracingManager.SetSpanAttribute(ctx, "error.service", metric.Service)
		s.tracingManager.SetSpanAttribute(ctx, "error.status_code", metric.StatusCode)

		err := s.metricsRepo.RecordErrorMetric(ctx, metric)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "record_error", "error_metric")
			return fmt.Errorf("failed to record error metric: %w", err)
		}

		// Записываем метрики в OpenTelemetry
		s.metricsCollector.RecordSystemError(ctx, metric.ErrorType, metric.Service)

		s.logger.Info("Error metric recorded",
			zap.String("error_type", metric.ErrorType),
			zap.String("service", metric.Service),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetErrorAnalytics получает аналитику ошибок
func (s *AnalyticsServiceImpl) GetErrorAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return s.executeWithTracingMetricResponse(ctx, "GetErrorAnalytics", func(ctx context.Context) (*domain.MetricResponse, error) {
		analytics, err := s.metricsRepo.GetErrorAnalytics(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "analytics_error", "error_analytics")
			return nil, fmt.Errorf("failed to get error analytics: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "error_analytics")
		return analytics, nil
	})
}

// GetErrorStats получает статистику ошибок
func (s *AnalyticsServiceImpl) GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return s.executeWithTracingMapInterface(ctx, "GetErrorStats", func(ctx context.Context) (map[string]interface{}, error) {
		stats, err := s.metricsRepo.GetErrorStats(ctx, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "stats_error", "error_stats")
			return nil, fmt.Errorf("failed to get error stats: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "error_stats")
		return stats, nil
	})
}

// GetTimeSeries получает временные ряды метрик
func (s *AnalyticsServiceImpl) GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error) {
	return s.executeWithTracingMetricSlice(ctx, "GetTimeSeries", func(ctx context.Context) ([]domain.Metric, error) {
		s.tracingManager.SetSpanAttribute(ctx, "timeseries.metric_name", metricName)
		s.tracingManager.SetSpanAttribute(ctx, "timeseries.interval", opts.Interval)

		_, err := s.metricsRepo.GetTimeSeries(ctx, metricName, opts)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "timeseries_error", "get_timeseries")
			return nil, fmt.Errorf("failed to get time series: %w", err)
		}

		s.metricsCollector.RecordMetricProcessed(ctx, "timeseries")
		metrics := []domain.Metric{}
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(metrics))

		return metrics, nil
	})
}

// CreateDashboard создает дашборд
func (s *AnalyticsServiceImpl) CreateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	return s.tracingManager.TraceDashboardRequest(ctx, dashboard.ID, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.name", dashboard.Name)
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.widgets_count", len(dashboard.Widgets))

		err := s.dashboardRepo.CreateDashboard(ctx, dashboard)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "create_error", "dashboard")
			return fmt.Errorf("failed to create dashboard: %w", err)
		}

		s.metricsCollector.RecordDashboardRequest(ctx, "create")
		s.logger.Info("Dashboard created",
			zap.String("name", dashboard.Name),
			zap.String("id", dashboard.ID),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// GetDashboard получает дашборд
func (s *AnalyticsServiceImpl) GetDashboard(ctx context.Context, id string) (*domain.DashboardConfig, error) {
	return s.executeWithTracingDashboardConfig(ctx, "GetDashboard", func(ctx context.Context) (*domain.DashboardConfig, error) {
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.id", id)

		dashboard, err := s.dashboardRepo.GetDashboard(ctx, id)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "get_error", "dashboard")
			return nil, fmt.Errorf("failed to get dashboard: %w", err)
		}

		s.metricsCollector.RecordDashboardRequest(ctx, "get")
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.name", dashboard.Name)
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.widgets_count", len(dashboard.Widgets))

		result := &domain.DashboardConfig{}
		return result, nil
	})
}

// UpdateDashboard обновляет дашборд
func (s *AnalyticsServiceImpl) UpdateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	return s.tracingManager.TraceDashboardRequest(ctx, dashboard.ID, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.name", dashboard.Name)
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.widgets_count", len(dashboard.Widgets))

		err := s.dashboardRepo.UpdateDashboard(ctx, dashboard)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "update_error", "dashboard")
			return fmt.Errorf("failed to update dashboard: %w", err)
		}

		s.metricsCollector.RecordDashboardRequest(ctx, "update")
		s.logger.Info("Dashboard updated",
			zap.String("id", dashboard.ID),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// DeleteDashboard удаляет дашборд
func (s *AnalyticsServiceImpl) DeleteDashboard(ctx context.Context, id string) error {
	return s.tracingManager.TraceDashboardRequest(ctx, id, func(ctx context.Context) error {
		s.tracingManager.SetSpanAttribute(ctx, "dashboard.id", id)

		err := s.dashboardRepo.DeleteDashboard(ctx, id)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "delete_error", "dashboard")
			return fmt.Errorf("failed to delete dashboard: %w", err)
		}

		s.metricsCollector.RecordDashboardRequest(ctx, "delete")
		s.logger.Info("Dashboard deleted",
			zap.String("id", id),
			zap.String("trace_id", s.tracingManager.GetTraceID(ctx)),
		)

		return nil
	})
}

// ListDashboards получает список дашбордов
func (s *AnalyticsServiceImpl) ListDashboards(ctx context.Context) ([]domain.DashboardConfig, error) {
	return s.executeWithTracingDashboardConfigs(ctx, "ListDashboards", func(ctx context.Context) ([]domain.DashboardConfig, error) {
		dashboards, err := s.dashboardRepo.ListDashboards(ctx)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "list_error", "dashboard")
			return nil, fmt.Errorf("failed to list dashboards: %w", err)
		}

		s.metricsCollector.RecordDashboardRequest(ctx, "list")
		s.tracingManager.SetSpanAttribute(ctx, "result.count", len(dashboards))

		return dashboards, nil
	})
}

// PredictBypassEffectiveness предсказывает эффективность обхода DPI
func (s *AnalyticsServiceImpl) PredictBypassEffectiveness(ctx context.Context, bypassType string, hoursAhead int) ([]domain.Metric, error) {
	return s.executeWithTracingMetricSlice(ctx, "PredictBypassEffectiveness", func(ctx context.Context) ([]domain.Metric, error) {
		s.tracingManager.SetSpanAttribute(ctx, "prediction.type", bypassType)
		s.tracingManager.SetSpanAttribute(ctx, "prediction.hours_ahead", hoursAhead)

		_, err := s.metricsRepo.PredictBypassEffectiveness(ctx, bypassType, hoursAhead)
		if err != nil {
			s.metricsCollector.RecordMetricError(ctx, "prediction_error", "bypass_effectiveness")
			return nil, fmt.Errorf("failed to predict bypass effectiveness: %w", err)
		}

		metrics := []domain.Metric{}
		s.metricsCollector.RecordPredictionGenerated(ctx, "bypass_effectiveness")
		s.tracingManager.SetSpanAttribute(ctx, "prediction.points", len(metrics))

		return metrics, nil
	})
}

// executeWithTracing выполняет функцию с трейсингом для []domain.ServerLoadMetric
func (s *AnalyticsServiceImpl) executeWithTracing(ctx context.Context, operation string, fn func(context.Context) ([]domain.ServerLoadMetric, error)) ([]domain.ServerLoadMetric, error) {
	var result []domain.ServerLoadMetric
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingMetric выполняет функцию с трейсингом для *domain.Metric
func (s *AnalyticsServiceImpl) executeWithTracingMetric(ctx context.Context, operation string, fn func(context.Context) (*domain.Metric, error)) (*domain.Metric, error) {
	var result *domain.Metric
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingMetrics выполняет функцию с трейсингом для []*domain.Metric
func (s *AnalyticsServiceImpl) executeWithTracingMetrics(ctx context.Context, operation string, fn func(context.Context) ([]*domain.Metric, int, error)) ([]*domain.Metric, int, error) {
	var result []*domain.Metric
	var total int
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, total, err = fn(ctx)
		return err
	})
	return result, total, err
}

// executeWithTracingTimeSeries выполняет функцию с трейсингом для []domain.TimeSeriesPoint
func (s *AnalyticsServiceImpl) executeWithTracingTimeSeries(ctx context.Context, operation string, fn func(context.Context) ([]domain.TimeSeriesPoint, error)) ([]domain.TimeSeriesPoint, error) {
	var result []domain.TimeSeriesPoint
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingStatistics выполняет функцию с трейсингом для []*domain.Statistics
func (s *AnalyticsServiceImpl) executeWithTracingStatistics(ctx context.Context, operation string, fn func(context.Context) ([]*domain.Statistics, error)) ([]*domain.Statistics, error) {
	var result []*domain.Statistics
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingSystemStats выполняет функцию с трейсингом для *domain.SystemStats
func (s *AnalyticsServiceImpl) executeWithTracingSystemStats(ctx context.Context, operation string, fn func(context.Context) (*domain.SystemStats, error)) (*domain.SystemStats, error) {
	var result *domain.SystemStats
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingUserStats выполняет функцию с трейсингом для *domain.UserStats
func (s *AnalyticsServiceImpl) executeWithTracingUserStats(ctx context.Context, operation string, fn func(context.Context) (*domain.UserStats, error)) (*domain.UserStats, error) {
	var result *domain.UserStats
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingDashboard выполняет функцию с трейсингом для *domain.DashboardData
func (s *AnalyticsServiceImpl) executeWithTracingDashboard(ctx context.Context, operation string, fn func(context.Context) (*domain.DashboardData, error)) (*domain.DashboardData, error) {
	var result *domain.DashboardData
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingPrediction выполняет функцию с трейсингом для []domain.PredictionPoint
func (s *AnalyticsServiceImpl) executeWithTracingPrediction(ctx context.Context, operation string, fn func(context.Context) ([]domain.PredictionPoint, error)) ([]domain.PredictionPoint, error) {
	var result []domain.PredictionPoint
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingMetricResponse выполняет функцию с трейсингом для *domain.MetricResponse
func (s *AnalyticsServiceImpl) executeWithTracingMetricResponse(ctx context.Context, operation string, fn func(context.Context) (*domain.MetricResponse, error)) (*domain.MetricResponse, error) {
	var result *domain.MetricResponse
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingMapInterface выполняет функцию с трейсингом для map[string]interface{}
func (s *AnalyticsServiceImpl) executeWithTracingMapInterface(ctx context.Context, operation string, fn func(context.Context) (map[string]interface{}, error)) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingMetricSlice выполняет функцию с трейсингом для []domain.Metric
func (s *AnalyticsServiceImpl) executeWithTracingMetricSlice(ctx context.Context, operation string, fn func(context.Context) ([]domain.Metric, error)) ([]domain.Metric, error) {
	var result []domain.Metric
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingDashboardConfig выполняет функцию с трейсингом для *domain.DashboardConfig
func (s *AnalyticsServiceImpl) executeWithTracingDashboardConfig(ctx context.Context, operation string, fn func(context.Context) (*domain.DashboardConfig, error)) (*domain.DashboardConfig, error) {
	var result *domain.DashboardConfig
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}

// executeWithTracingDashboardConfigs выполняет функцию с трейсингом для []domain.DashboardConfig
func (s *AnalyticsServiceImpl) executeWithTracingDashboardConfigs(ctx context.Context, operation string, fn func(context.Context) ([]domain.DashboardConfig, error)) ([]domain.DashboardConfig, error) {
	var result []domain.DashboardConfig
	err := s.tracingManager.TraceAnalyticsOperation(ctx, operation, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	})
	return result, err
}
