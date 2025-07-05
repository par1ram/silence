package database

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// SaveErrorMetric сохраняет метрику ошибки
func (r *InfluxDBRepository) SaveErrorMetric(ctx context.Context, metric domain.ErrorMetric) error {
	point := influxdb2.NewPoint(
		"error_metrics",
		map[string]string{
			"error_type": metric.ErrorType,
			"service":    metric.Service,
			"user_id":    metric.UserID,
			"server_id":  metric.ServerID,
		},
		map[string]interface{}{
			"status_code": metric.StatusCode,
			"description": metric.Description,
			"value":       metric.Value,
		},
		metric.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, point)
}

// GetErrorMetrics получает метрики ошибок
func (r *InfluxDBRepository) GetErrorMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	query := r.buildQuery("error_metrics", opts)
	return r.executeQuery(ctx, query, opts)
}

// GetErrorStats получает статистику ошибок
func (r *InfluxDBRepository) GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_errors": 0,
		"error_rate":   0,
		"most_common":  "",
	}, nil
}
