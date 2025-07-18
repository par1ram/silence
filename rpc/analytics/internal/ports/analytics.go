package ports

import (
	"context"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// MetricsRepository интерфейс для работы с метриками
type MetricsRepository interface {
	// Сохранение метрик
	SaveConnectionMetric(ctx context.Context, metric domain.ConnectionMetric) error
	SaveBypassEffectivenessMetric(ctx context.Context, metric domain.BypassEffectivenessMetric) error
	SaveUserActivityMetric(ctx context.Context, metric domain.UserActivityMetric) error
	SaveServerLoadMetric(ctx context.Context, metric domain.ServerLoadMetric) error
	SaveErrorMetric(ctx context.Context, metric domain.ErrorMetric) error

	// Запросы метрик
	GetConnectionMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetBypassEffectivenessMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetUserActivityMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetServerLoadMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetErrorMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)

	// Агрегированные запросы
	GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)
	GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)
	GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)
	GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)
	GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)

	// Временные серии
	GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error)

	// Дополнительные методы для gRPC handler
	SaveMetric(ctx context.Context, metric *domain.Metric) error
	GetMetrics(ctx context.Context, filters *domain.MetricFilters) ([]*domain.Metric, int, error)
	GetMetricsHistory(ctx context.Context, req *domain.MetricHistoryRequest) ([]domain.TimeSeriesPoint, error)
	GetStatistics(ctx context.Context, req *domain.StatisticsRequest) ([]*domain.Statistics, error)
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)
	GetUserStats(ctx context.Context, req *domain.UserStatsRequest) (*domain.UserStats, error)
	PredictLoad(ctx context.Context, req *domain.PredictionRequest) ([]domain.PredictionPoint, error)
	PredictTrend(ctx context.Context, req *domain.TrendRequest) ([]domain.PredictionPoint, error)
	PredictBypassEffectiveness(ctx context.Context, bypassType string, hours int) ([]domain.Metric, error)

	// Методы для записи метрик
	RecordConnectionMetric(ctx context.Context, metric domain.ConnectionMetric) error
	RecordBypassEffectivenessMetric(ctx context.Context, metric domain.BypassEffectivenessMetric) error
	RecordUserActivityMetric(ctx context.Context, metric domain.UserActivityMetric) error
	RecordServerLoadMetric(ctx context.Context, metric domain.ServerLoadMetric) error
	RecordErrorMetric(ctx context.Context, metric domain.ErrorMetric) error

	// Методы для получения аналитики
	GetConnectionAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetBypassEffectivenessAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetUserActivityAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetServerLoadAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetErrorAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
}

// DashboardRepository интерфейс для работы с дашбордами
type DashboardRepository interface {
	CreateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error
	GetDashboard(ctx context.Context, id string) (*domain.DashboardConfig, error)
	UpdateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error
	DeleteDashboard(ctx context.Context, id string) error
	ListDashboards(ctx context.Context) ([]domain.DashboardConfig, error)
	GetDashboardData(ctx context.Context, timeRange string) (*domain.DashboardData, error)
}

// AnalyticsService основной сервис аналитики
type AnalyticsService interface {
	// Метрики подключений
	RecordConnection(ctx context.Context, metric domain.ConnectionMetric) error
	GetConnectionAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)

	// Метрики обхода DPI
	RecordBypassEffectiveness(ctx context.Context, metric domain.BypassEffectivenessMetric) error
	GetBypassEffectivenessAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)

	// Метрики активности пользователей
	RecordUserActivity(ctx context.Context, metric domain.UserActivityMetric) error
	GetUserActivityAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)

	// Метрики нагрузки серверов
	RecordServerLoad(ctx context.Context, metric domain.ServerLoadMetric) error
	GetServerLoadAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)
	GetServerLoadMetrics(ctx context.Context, start, end time.Time) ([]domain.ServerLoadMetric, error)

	// Метрики ошибок
	RecordError(ctx context.Context, metric domain.ErrorMetric) error
	GetErrorAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error)
	GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error)

	// Временные серии
	GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error)

	// Дашборды
	CreateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error
	GetDashboard(ctx context.Context, id string) (*domain.DashboardConfig, error)
	UpdateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error
	DeleteDashboard(ctx context.Context, id string) error
	ListDashboards(ctx context.Context) ([]domain.DashboardConfig, error)

	// Методы для gRPC handler
	CollectMetric(ctx context.Context, metric *domain.Metric) (*domain.Metric, error)
	GetMetrics(ctx context.Context, filters *domain.MetricFilters) ([]*domain.Metric, int, error)
	GetMetricsHistory(ctx context.Context, req *domain.MetricHistoryRequest) ([]domain.TimeSeriesPoint, error)
	GetStatistics(ctx context.Context, req *domain.StatisticsRequest) ([]*domain.Statistics, error)
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)
	GetUserStats(ctx context.Context, req *domain.UserStatsRequest) (*domain.UserStats, error)
	GetDashboardData(ctx context.Context, timeRange string) (*domain.DashboardData, error)
	PredictLoad(ctx context.Context, req *domain.PredictionRequest) ([]domain.PredictionPoint, error)
	PredictTrend(ctx context.Context, req *domain.TrendRequest) ([]domain.PredictionPoint, error)
	PredictBypassEffectiveness(ctx context.Context, bypassType string, hours int) ([]domain.Metric, error)
}

// MetricsCollector интерфейс для сбора метрик
type MetricsCollector interface {
	// Сбор метрик из других сервисов
	CollectConnectionMetrics(ctx context.Context) error
	CollectBypassEffectivenessMetrics(ctx context.Context) error
	CollectUserActivityMetrics(ctx context.Context) error
	CollectServerLoadMetrics(ctx context.Context) error
	CollectErrorMetrics(ctx context.Context) error

	// Периодический сбор
	StartPeriodicCollection(ctx context.Context) error
	StopPeriodicCollection(ctx context.Context) error
}

// AlertService интерфейс для алертов
type AlertService interface {
	// Управление правилами алертов
	CreateAlertRule(ctx context.Context, rule domain.AlertRule) error
	GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error)
	UpdateAlertRule(ctx context.Context, rule domain.AlertRule) error
	DeleteAlertRule(ctx context.Context, id string) error
	ListAlertRules(ctx context.Context) ([]domain.AlertRule, error)

	// Оценка алертов
	EvaluateAlerts(ctx context.Context) error

	// Управление уведомлениями
	GetAlertHistory(ctx context.Context, ruleID string, limit int) ([]domain.Alert, error)
	AcknowledgeAlert(ctx context.Context, alertID string) error
	ResolveAlert(ctx context.Context, alertID string) error
}
