package database

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
)

// InfluxDBRepository реализация репозитория метрик с InfluxDB
type InfluxDBRepository struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
	bucket   string
	org      string
	logger   *zap.Logger
}

// NewInfluxDBRepository создает новый репозиторий InfluxDB
func NewInfluxDBRepository(url, token, org, bucket string, logger *zap.Logger) (ports.MetricsRepository, error) {
	client := influxdb2.NewClient(url, token)

	// Проверка соединения
	_, err := client.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping InfluxDB: %w", err)
	}

	writeAPI := client.WriteAPIBlocking(org, bucket)
	queryAPI := client.QueryAPI(org)

	return &InfluxDBRepository{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		bucket:   bucket,
		org:      org,
		logger:   logger,
	}, nil
}

// Close закрывает соединение с InfluxDB
func (r *InfluxDBRepository) Close() {
	r.client.Close()
}

// SaveConnectionMetric сохраняет метрику подключения
func (r *InfluxDBRepository) SaveConnectionMetric(ctx context.Context, metric domain.ConnectionMetric) error {
	point := influxdb2.NewPoint(
		"connection_metrics",
		map[string]string{
			"user_id":     metric.UserID,
			"server_id":   metric.ServerID,
			"protocol":    metric.Protocol,
			"bypass_type": metric.BypassType,
			"region":      metric.Region,
		},
		map[string]interface{}{
			"duration_ms": metric.Duration,
			"bytes_in":    metric.BytesIn,
			"bytes_out":   metric.BytesOut,
			"value":       metric.Value,
		},
		metric.Timestamp,
	)

	return r.writeAPI.WritePoint(ctx, point)
}

// GetConnectionMetrics получает метрики подключений
func (r *InfluxDBRepository) GetConnectionMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	query := r.buildQuery("connection_metrics", opts)
	return r.executeQuery(ctx, query, opts)
}

// GetConnectionStats получает статистику подключений
func (r *InfluxDBRepository) GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	// TODO: Реализовать сложные агрегации
	return map[string]interface{}{
		"total_connections": 0,
		"avg_duration":      0,
		"total_bytes":       0,
	}, nil
}
