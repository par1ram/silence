package database

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	write "github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// SaveBypassEffectivenessMetric сохраняет метрику эффективности обхода DPI
func (r *InfluxDBRepository) SaveBypassEffectivenessMetric(ctx context.Context, metric domain.BypassEffectivenessMetric) error {
	point := r.newBypassEffectivenessPoint(metric)
	return r.writeAPI.WritePoint(ctx, point)
}

func (r *InfluxDBRepository) newBypassEffectivenessPoint(metric domain.BypassEffectivenessMetric) *write.Point {
	return influxdb2.NewPoint(
		"bypass_effectiveness_metrics",
		map[string]string{
			"bypass_type": metric.BypassType,
		},
		map[string]interface{}{
			"success_rate":    metric.SuccessRate,
			"latency_ms":      metric.Latency,
			"throughput_mbps": metric.Throughput,
			"blocked_count":   metric.BlockedCount,
			"total_attempts":  metric.TotalAttempts,
			"value":           metric.Value,
		},
		metric.Timestamp,
	)
}

// GetBypassEffectivenessMetrics получает метрики эффективности обхода DPI
func (r *InfluxDBRepository) GetBypassEffectivenessMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	query := r.buildQuery("bypass_effectiveness_metrics", opts)
	return r.executeQuery(ctx, query, opts)
}

// GetBypassEffectivenessStats получает статистику эффективности обхода DPI
func (r *InfluxDBRepository) GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{
		"avg_success_rate": 0,
		"avg_latency":      0,
		"total_attempts":   0,
	}, nil
}
