package database

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// buildQuery строит Flux запрос
func (r *InfluxDBRepository) buildQuery(measurement string, opts domain.QueryOptions) string {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "%s")
	`, r.bucket, opts.TimeRange.Start.Format(time.RFC3339), opts.TimeRange.End.Format(time.RFC3339), measurement)

	for key, value := range opts.Filters {
		query += fmt.Sprintf(`|> filter(fn: (r) => r["%s"] == "%s")`, key, value)
	}

	if len(opts.GroupBy) > 0 {
		query += fmt.Sprintf(`|> group(columns: ["%s"])`, opts.GroupBy[0])
	}

	switch opts.Aggregation {
	case domain.AggregationSum:
		query += `|> sum()`
	case domain.AggregationAvg:
		query += `|> mean()`
	case domain.AggregationMin:
		query += `|> min()`
	case domain.AggregationMax:
		query += `|> max()`
	case domain.AggregationCount:
		query += `|> count()`
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(`|> limit(n: %d)`, opts.Limit)
	}

	return query
}

// executeQuery выполняет Flux запрос
func (r *InfluxDBRepository) executeQuery(ctx context.Context, query string, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer result.Close()

	var metrics []domain.Metric
	for result.Next() {
		record := result.Record()
		metric := domain.Metric{
			Name:      record.Measurement(),
			Timestamp: record.Time(),
			Labels:    make(map[string]string),
		}
		for key, value := range record.Values() {
			if key != "_time" && key != "_value" && key != "_field" && key != "_measurement" {
				metric.Labels[key] = fmt.Sprintf("%v", value)
			}
		}
		if value, ok := record.Value().(float64); ok {
			metric.Value = value
		}
		metrics = append(metrics, metric)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("error iterating results: %w", result.Err())
	}
	return &domain.MetricResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
		HasMore: false,
	}, nil
}

// GetTimeSeries получает временные серии
func (r *InfluxDBRepository) GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
	`, r.bucket, opts.TimeRange.Start.Format(time.RFC3339), opts.TimeRange.End.Format(time.RFC3339), metricName, opts.Interval)

	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute time series query: %w", err)
	}
	defer result.Close()

	var metrics []domain.Metric
	for result.Next() {
		record := result.Record()
		metric := domain.Metric{
			Name:      record.Measurement(),
			Timestamp: record.Time(),
		}
		if value, ok := record.Value().(float64); ok {
			metric.Value = value
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
