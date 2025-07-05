package database

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// SaveServerLoadMetric сохраняет метрику нагрузки сервера
func (r *InfluxDBRepository) SaveServerLoadMetric(ctx context.Context, metric domain.ServerLoadMetric) error {
	point := influxdb2.NewPoint(
		"server_load_metrics",
		map[string]string{
			"server_id": metric.ServerID,
			"region":    metric.Region,
		},
		map[string]interface{}{
			"cpu_usage_percent":    metric.CPUUsage,
			"memory_usage_percent": metric.MemoryUsage,
			"network_in_mbps":      metric.NetworkIn,
			"network_out_mbps":     metric.NetworkOut,
			"active_connections":   metric.Connections,
			"value":                metric.Value,
		},
		metric.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, point)
}

// GetServerLoadMetrics получает метрики нагрузки серверов
func (r *InfluxDBRepository) GetServerLoadMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	query := r.buildQuery("server_load_metrics", opts)
	return r.executeQuery(ctx, query, opts)
}

// GetServerLoadStats получает статистику нагрузки серверов
func (r *InfluxDBRepository) GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{
		"avg_cpu_usage":     0,
		"avg_memory_usage":  0,
		"total_connections": 0,
	}, nil
}
