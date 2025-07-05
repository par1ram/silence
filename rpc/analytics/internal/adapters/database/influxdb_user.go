package database

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// SaveUserActivityMetric сохраняет метрику активности пользователя
func (r *InfluxDBRepository) SaveUserActivityMetric(ctx context.Context, metric domain.UserActivityMetric) error {
	point := influxdb2.NewPoint(
		"user_activity_metrics",
		map[string]string{
			"user_id": metric.UserID,
		},
		map[string]interface{}{
			"session_count":      metric.SessionCount,
			"total_time_minutes": metric.TotalTime,
			"data_usage_mb":      metric.DataUsage,
			"login_count":        metric.LoginCount,
			"value":              metric.Value,
		},
		metric.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, point)
}

// GetUserActivityMetrics получает метрики активности пользователей
func (r *InfluxDBRepository) GetUserActivityMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	query := r.buildQuery("user_activity_metrics", opts)
	return r.executeQuery(ctx, query, opts)
}

// GetUserActivityStats получает статистику активности пользователей
func (r *InfluxDBRepository) GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{
		"active_users":     0,
		"total_sessions":   0,
		"avg_session_time": 0,
	}, nil
}
