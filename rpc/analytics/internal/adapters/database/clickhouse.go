package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
)

// ClickHouseRepository реализация репозитория метрик с ClickHouse
type ClickHouseRepository struct {
	conn   driver.Conn
	logger *zap.Logger
}

// NewClickHouseRepository создает новый репозиторий ClickHouse
func NewClickHouseRepository(host string, port int, database, username, password string, logger *zap.Logger) (ports.MetricsRepository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Проверка соединения
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseRepository{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close закрывает соединение с ClickHouse
func (r *ClickHouseRepository) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

// SaveMetric сохраняет базовую метрику
func (r *ClickHouseRepository) SaveMetric(ctx context.Context, metric *domain.Metric) error {
	query := `INSERT INTO metrics (id, service_name, metric_type, metric_name, metric_value, labels, timestamp, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	labels := make(map[string]string)
	if metric.Tags != nil {
		for k, v := range metric.Tags {
			labels[k] = v
		}
	}

	return r.conn.Exec(ctx, query,
		uuid.New().String(),
		"analytics",
		metric.Type,
		metric.Name,
		metric.Value,
		labels,
		metric.Timestamp,
		time.Now(),
	)
}

// GetMetrics получает метрики с фильтрами
func (r *ClickHouseRepository) GetMetrics(ctx context.Context, filters *domain.MetricFilters) ([]*domain.Metric, int, error) {
	var conditions []string
	var args []interface{}

	if filters.Name != "" {
		conditions = append(conditions, "metric_name = ?")
		args = append(args, filters.Name)
	}

	if !filters.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filters.StartTime)
	}

	if !filters.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filters.EndTime)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество записей
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM metrics %s", whereClause)
	var total int
	if err := r.conn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Получаем данные с пагинацией
	query := fmt.Sprintf(`
		SELECT service_name, metric_type, metric_name, metric_value, labels, timestamp, created_at
		FROM metrics
		%s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*domain.Metric
	for rows.Next() {
		var m domain.Metric
		var labels map[string]string
		var serviceName, metricType string

		if err := rows.Scan(&serviceName, &metricType, &m.Name, &m.Value, &labels, &m.Timestamp, &m.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("failed to scan metric row: %w", err)
		}

		m.Type = metricType
		m.Tags = labels
		metrics = append(metrics, &m)
	}

	return metrics, total, nil
}

// GetMetricsHistory получает историю метрик
func (r *ClickHouseRepository) GetMetricsHistory(ctx context.Context, req *domain.MetricHistoryRequest) ([]domain.TimeSeriesPoint, error) {
	query := `
		SELECT
			toStartOfInterval(timestamp, INTERVAL ? minute) as time_bucket,
			AVG(metric_value) as avg_value,
			MIN(metric_value) as min_value,
			MAX(metric_value) as max_value,
			COUNT(*) as count
		FROM metrics
		WHERE service_name = ?
		AND metric_name = ?
		AND timestamp >= ?
		AND timestamp <= ?
		GROUP BY time_bucket
		ORDER BY time_bucket ASC`

	interval := 5 // 5 минут по умолчанию

	rows, err := r.conn.Query(ctx, query,
		interval,
		"analytics",
		req.Name,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics history: %w", err)
	}
	defer rows.Close()

	var points []domain.TimeSeriesPoint
	for rows.Next() {
		var point domain.TimeSeriesPoint
		var timeBucket time.Time
		var avgValue, minValue, maxValue float64
		var count int64

		if err := rows.Scan(&timeBucket, &avgValue, &minValue, &maxValue, &count); err != nil {
			return nil, fmt.Errorf("failed to scan time series point: %w", err)
		}

		point.Timestamp = timeBucket
		point.Value = avgValue
		points = append(points, point)
	}

	return points, nil
}

// GetStatistics получает статистику
func (r *ClickHouseRepository) GetStatistics(ctx context.Context, req *domain.StatisticsRequest) ([]*domain.Statistics, error) {
	query := `
		SELECT
			service_name,
			metric_type,
			COUNT(*) as total_count,
			AVG(metric_value) as avg_value,
			MIN(metric_value) as min_value,
			MAX(metric_value) as max_value,
			stddevPop(metric_value) as std_dev
		FROM metrics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY service_name, metric_type
		ORDER BY service_name, metric_type`

	rows, err := r.conn.Query(ctx, query, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}
	defer rows.Close()

	var stats []*domain.Statistics
	for rows.Next() {
		var stat domain.Statistics
		var serviceName, metricType string
		var totalCount int64
		var avgValue, minValue, maxValue, stdDev float64

		if err := rows.Scan(&serviceName, &metricType, &totalCount, &avgValue, &minValue, &maxValue, &stdDev); err != nil {
			return nil, fmt.Errorf("failed to scan statistics row: %w", err)
		}

		stat.Name = serviceName
		stat.Type = metricType
		stat.Value = avgValue

		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetSystemStats получает системную статистику
func (r *ClickHouseRepository) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	query := `
		SELECT
			COUNT(*) as total_metrics,
			COUNT(DISTINCT service_name) as total_services,
			COUNT(DISTINCT metric_type) as total_metric_types,
			MIN(timestamp) as earliest_metric,
			MAX(timestamp) as latest_metric
		FROM metrics`

	var stats domain.SystemStats
	var totalMetrics, totalServices, totalMetricTypes int64
	var earliestMetric, latestMetric time.Time

	err := r.conn.QueryRow(ctx, query).Scan(
		&totalMetrics,
		&totalServices,
		&totalMetricTypes,
		&earliestMetric,
		&latestMetric,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query system stats: %w", err)
	}

	stats.TotalUsers = totalMetrics
	stats.ActiveUsers = totalServices
	stats.TotalConnections = totalMetricTypes
	stats.LastUpdated = latestMetric

	return &stats, nil
}

// GetUserStats получает статистику пользователя
func (r *ClickHouseRepository) GetUserStats(ctx context.Context, req *domain.UserStatsRequest) (*domain.UserStats, error) {
	// Получаем статистику из user_analytics таблицы
	query := `
		SELECT
			COUNT(*) as total_actions,
			COUNT(DISTINCT action) as unique_actions,
			COUNT(DISTINCT toDate(timestamp)) as active_days,
			MIN(timestamp) as first_action,
			MAX(timestamp) as last_action
		FROM user_analytics
		WHERE user_id = ?
		AND timestamp >= ?
		AND timestamp <= ?`

	var stats domain.UserStats
	var totalActions, uniqueActions, activeDays int64
	var firstAction, lastAction time.Time

	err := r.conn.QueryRow(ctx, query,
		req.UserID,
		req.StartTime,
		req.EndTime,
	).Scan(
		&totalActions,
		&uniqueActions,
		&activeDays,
		&firstAction,
		&lastAction,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query user stats: %w", err)
	}

	stats.UserID = req.UserID
	stats.TotalConnections = totalActions
	stats.TotalDataTransferred = uniqueActions
	stats.TotalSessionTime = activeDays
	stats.FirstConnection = firstAction
	stats.LastConnection = lastAction

	return &stats, nil
}

// PredictLoad предсказывает нагрузку (простая реализация)
func (r *ClickHouseRepository) PredictLoad(ctx context.Context, req *domain.PredictionRequest) ([]domain.PredictionPoint, error) {
	// Получаем исторические данные для предсказания
	query := `
		SELECT
			toStartOfHour(timestamp) as hour,
			AVG(metric_value) as avg_value
		FROM metrics
		WHERE service_name = ?
		AND metric_name = 'cpu_usage'
		AND timestamp >= ?
		AND timestamp <= ?
		GROUP BY hour
		ORDER BY hour DESC
		LIMIT 24`

	startTime := time.Now()
	rows, err := r.conn.Query(ctx, query,
		"analytics",
		startTime.Add(-24*time.Hour),
		startTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical data: %w", err)
	}
	defer rows.Close()

	var historicalData []float64
	for rows.Next() {
		var hour time.Time
		var avgValue float64
		if err := rows.Scan(&hour, &avgValue); err != nil {
			return nil, fmt.Errorf("failed to scan historical data: %w", err)
		}
		historicalData = append(historicalData, avgValue)
	}

	// Простое предсказание на основе среднего значения
	var predictions []domain.PredictionPoint
	if len(historicalData) > 0 {
		var sum float64
		for _, value := range historicalData {
			sum += value
		}
		avgValue := sum / float64(len(historicalData))

		// Генерируем предсказания на следующие часы
		for i := 0; i < req.HoursAhead; i++ {
			prediction := domain.PredictionPoint{
				Timestamp:      startTime.Add(time.Duration(i) * time.Hour),
				PredictedValue: avgValue,
				Confidence:     0.8, // Простая оценка уверенности
			}
			predictions = append(predictions, prediction)
		}
	}

	return predictions, nil
}

// PredictTrend предсказывает тренд
func (r *ClickHouseRepository) PredictTrend(ctx context.Context, req *domain.TrendRequest) ([]domain.PredictionPoint, error) {
	// Получаем данные для анализа тренда
	query := `
		SELECT
			toStartOfDay(timestamp) as day,
			AVG(metric_value) as avg_value
		FROM metrics
		WHERE service_name = ?
		AND metric_name = ?
		AND timestamp >= ?
		AND timestamp <= ?
		GROUP BY day
		ORDER BY day ASC`

	startTime := time.Now()
	rows, err := r.conn.Query(ctx, query,
		"analytics",
		req.MetricName,
		startTime.Add(-time.Duration(req.DaysAhead)*24*time.Hour),
		startTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend data: %w", err)
	}
	defer rows.Close()

	var dataPoints []domain.PredictionPoint
	for rows.Next() {
		var day time.Time
		var avgValue float64
		if err := rows.Scan(&day, &avgValue); err != nil {
			return nil, fmt.Errorf("failed to scan trend data: %w", err)
		}
		dataPoints = append(dataPoints, domain.PredictionPoint{
			Timestamp:      day,
			PredictedValue: avgValue,
		})
	}

	// Простой линейный тренд
	var predictions []domain.PredictionPoint
	if len(dataPoints) >= 2 {
		// Вычисляем простой линейный тренд
		firstPoint := dataPoints[0]
		lastPoint := dataPoints[len(dataPoints)-1]

		timeDiff := lastPoint.Timestamp.Sub(firstPoint.Timestamp).Hours()
		valueDiff := lastPoint.PredictedValue - firstPoint.PredictedValue

		if timeDiff > 0 {
			slope := valueDiff / timeDiff

			// Генерируем предсказания
			for i := 0; i < req.DaysAhead; i++ {
				futureTime := startTime.Add(time.Duration(i) * 24 * time.Hour)
				hoursFromLast := futureTime.Sub(lastPoint.Timestamp).Hours()
				predictedValue := lastPoint.PredictedValue + slope*hoursFromLast

				prediction := domain.PredictionPoint{
					Timestamp:      futureTime,
					PredictedValue: predictedValue,
					Confidence:     0.7,
				}
				predictions = append(predictions, prediction)
			}
		}
	}

	return predictions, nil
}

// Методы для совместимости с интерфейсом (пока заглушки)
func (r *ClickHouseRepository) SaveConnectionMetric(ctx context.Context, metric domain.ConnectionMetric) error {
	return r.insertEvent(ctx, "connection_metrics", map[string]interface{}{
		"user_id":     metric.UserID,
		"server_id":   metric.ServerID,
		"protocol":    metric.Protocol,
		"bypass_type": metric.BypassType,
		"region":      metric.Region,
		"duration":    metric.Duration,
		"bytes_in":    metric.BytesIn,
		"bytes_out":   metric.BytesOut,
		"value":       metric.Value,
		"timestamp":   metric.Timestamp,
	})
}

func (r *ClickHouseRepository) GetConnectionMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return &domain.MetricResponse{Metrics: []domain.Metric{}, Total: 0, HasMore: false}, nil
}

func (r *ClickHouseRepository) GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *ClickHouseRepository) SaveBypassEffectivenessMetric(ctx context.Context, metric domain.BypassEffectivenessMetric) error {
	return r.insertEvent(ctx, "bypass_effectiveness", map[string]interface{}{
		"bypass_type":    metric.BypassType,
		"success_rate":   metric.SuccessRate,
		"latency":        metric.Latency,
		"throughput":     metric.Throughput,
		"blocked_count":  metric.BlockedCount,
		"total_attempts": metric.TotalAttempts,
		"value":          metric.Value,
		"timestamp":      metric.Timestamp,
	})
}

func (r *ClickHouseRepository) GetBypassEffectivenessMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return &domain.MetricResponse{Metrics: []domain.Metric{}, Total: 0, HasMore: false}, nil
}

func (r *ClickHouseRepository) GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *ClickHouseRepository) PredictBypassEffectiveness(ctx context.Context, bypassType string, hours int) ([]domain.Metric, error) {
	return []domain.Metric{}, nil
}

// Вспомогательные методы
func (r *ClickHouseRepository) insertEvent(ctx context.Context, tableName string, data map[string]interface{}) error {
	// Простая реализация вставки событий
	query := fmt.Sprintf("INSERT INTO %s (id, service_name, properties, timestamp, created_at) VALUES (?, ?, ?, ?, ?)", tableName)

	properties := make(map[string]string)
	for k, v := range data {
		if k != "timestamp" {
			properties[k] = fmt.Sprintf("%v", v)
		}
	}

	timestamp := time.Now()
	if ts, ok := data["timestamp"].(time.Time); ok {
		timestamp = ts
	}

	return r.conn.Exec(ctx, query,
		uuid.New().String(),
		"analytics",
		properties,
		timestamp,
		time.Now(),
	)
}

// Остальные методы для совместимости с интерфейсом
func (r *ClickHouseRepository) SaveUserActivityMetric(ctx context.Context, metric domain.UserActivityMetric) error {
	return r.insertEvent(ctx, "user_analytics", map[string]interface{}{
		"user_id":       metric.UserID,
		"action":        "activity",
		"session_count": metric.SessionCount,
		"total_time":    metric.TotalTime,
		"data_usage":    metric.DataUsage,
		"login_count":   metric.LoginCount,
		"value":         metric.Value,
		"timestamp":     metric.Timestamp,
	})
}

func (r *ClickHouseRepository) GetUserActivityMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return &domain.MetricResponse{Metrics: []domain.Metric{}, Total: 0, HasMore: false}, nil
}

func (r *ClickHouseRepository) GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *ClickHouseRepository) SaveServerLoadMetric(ctx context.Context, metric domain.ServerLoadMetric) error {
	return r.insertEvent(ctx, "server_stats", map[string]interface{}{
		"server_id":    metric.ServerID,
		"region":       metric.Region,
		"cpu_usage":    metric.CPUUsage,
		"memory_usage": metric.MemoryUsage,
		"network_in":   metric.NetworkIn,
		"network_out":  metric.NetworkOut,
		"connections":  metric.Connections,
		"value":        metric.Value,
		"timestamp":    metric.Timestamp,
	})
}

func (r *ClickHouseRepository) GetServerLoadMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return &domain.MetricResponse{Metrics: []domain.Metric{}, Total: 0, HasMore: false}, nil
}

func (r *ClickHouseRepository) GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *ClickHouseRepository) SaveErrorMetric(ctx context.Context, metric domain.ErrorMetric) error {
	return r.insertEvent(ctx, "alerts", map[string]interface{}{
		"alert_type":   "error",
		"severity":     "high",
		"service_name": metric.Service,
		"message":      metric.Description,
		"error_type":   metric.ErrorType,
		"user_id":      metric.UserID,
		"server_id":    metric.ServerID,
		"status_code":  metric.StatusCode,
		"value":        metric.Value,
		"timestamp":    metric.Timestamp,
	})
}

func (r *ClickHouseRepository) GetErrorMetrics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return &domain.MetricResponse{Metrics: []domain.Metric{}, Total: 0, HasMore: false}, nil
}

func (r *ClickHouseRepository) GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *ClickHouseRepository) GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error) {
	return []domain.Metric{}, nil
}

// Методы для совместимости с дополнительными интерфейсами
func (r *ClickHouseRepository) RecordConnectionMetric(ctx context.Context, metric domain.ConnectionMetric) error {
	return r.SaveConnectionMetric(ctx, metric)
}

func (r *ClickHouseRepository) RecordBypassEffectivenessMetric(ctx context.Context, metric domain.BypassEffectivenessMetric) error {
	return r.SaveBypassEffectivenessMetric(ctx, metric)
}

func (r *ClickHouseRepository) RecordUserActivityMetric(ctx context.Context, metric domain.UserActivityMetric) error {
	return r.SaveUserActivityMetric(ctx, metric)
}

func (r *ClickHouseRepository) RecordServerLoadMetric(ctx context.Context, metric domain.ServerLoadMetric) error {
	return r.SaveServerLoadMetric(ctx, metric)
}

func (r *ClickHouseRepository) RecordErrorMetric(ctx context.Context, metric domain.ErrorMetric) error {
	return r.SaveErrorMetric(ctx, metric)
}

func (r *ClickHouseRepository) GetConnectionAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return r.GetConnectionMetrics(ctx, opts)
}

func (r *ClickHouseRepository) GetBypassEffectivenessAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return r.GetBypassEffectivenessMetrics(ctx, opts)
}

func (r *ClickHouseRepository) GetUserActivityAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return r.GetUserActivityMetrics(ctx, opts)
}

func (r *ClickHouseRepository) GetServerLoadAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return r.GetServerLoadMetrics(ctx, opts)
}

func (r *ClickHouseRepository) GetErrorAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	return r.GetErrorMetrics(ctx, opts)
}
