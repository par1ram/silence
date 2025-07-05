package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

// Методы для работы с метриками

func (s *AnalyticsServiceImpl) RecordConnection(ctx context.Context, metric domain.ConnectionMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if err := s.metricsRepo.SaveConnectionMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save connection metric",
			zap.String("error", err.Error()),
			zap.String("user_id", metric.UserID),
			zap.String("server_id", metric.ServerID),
		)
		return fmt.Errorf("failed to save connection metric: %w", err)
	}

	s.logger.Debug("Connection metric recorded",
		zap.String("user_id", metric.UserID),
		zap.String("server_id", metric.ServerID),
		zap.Int64("duration", metric.Duration),
	)

	return nil
}

func (s *AnalyticsServiceImpl) GetConnectionAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	response, err := s.metricsRepo.GetConnectionMetrics(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get connection analytics", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get connection analytics: %w", err)
	}

	return response, nil
}

func (s *AnalyticsServiceImpl) GetConnectionStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	stats, err := s.metricsRepo.GetConnectionStats(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get connection stats", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get connection stats: %w", err)
	}

	return stats, nil
}

func (s *AnalyticsServiceImpl) RecordBypassEffectiveness(ctx context.Context, metric domain.BypassEffectivenessMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if err := s.metricsRepo.SaveBypassEffectivenessMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save bypass effectiveness metric",
			zap.String("error", err.Error()),
			zap.String("bypass_type", metric.BypassType),
		)
		return fmt.Errorf("failed to save bypass effectiveness metric: %w", err)
	}

	s.logger.Debug("Bypass effectiveness metric recorded",
		zap.String("bypass_type", metric.BypassType),
		zap.Float64("success_rate", metric.SuccessRate),
	)

	return nil
}

func (s *AnalyticsServiceImpl) GetBypassEffectivenessAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	response, err := s.metricsRepo.GetBypassEffectivenessMetrics(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get bypass effectiveness analytics", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get bypass effectiveness analytics: %w", err)
	}

	return response, nil
}

func (s *AnalyticsServiceImpl) GetBypassEffectivenessStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	stats, err := s.metricsRepo.GetBypassEffectivenessStats(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get bypass effectiveness stats", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get bypass effectiveness stats: %w", err)
	}

	return stats, nil
}

func (s *AnalyticsServiceImpl) RecordUserActivity(ctx context.Context, metric domain.UserActivityMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if err := s.metricsRepo.SaveUserActivityMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save user activity metric",
			zap.String("error", err.Error()),
			zap.String("user_id", metric.UserID),
		)
		return fmt.Errorf("failed to save user activity metric: %w", err)
	}

	s.logger.Debug("User activity metric recorded",
		zap.String("user_id", metric.UserID),
		zap.Int64("session_count", metric.SessionCount),
	)

	return nil
}

func (s *AnalyticsServiceImpl) GetUserActivityAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	response, err := s.metricsRepo.GetUserActivityMetrics(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get user activity analytics", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get user activity analytics: %w", err)
	}

	return response, nil
}

func (s *AnalyticsServiceImpl) GetUserActivityStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	stats, err := s.metricsRepo.GetUserActivityStats(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get user activity stats", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get user activity stats: %w", err)
	}

	return stats, nil
}

func (s *AnalyticsServiceImpl) RecordServerLoad(ctx context.Context, metric domain.ServerLoadMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if err := s.metricsRepo.SaveServerLoadMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save server load metric",
			zap.String("error", err.Error()),
			zap.String("server_id", metric.ServerID),
		)
		return fmt.Errorf("failed to save server load metric: %w", err)
	}

	s.logger.Debug("Server load metric recorded",
		zap.String("server_id", metric.ServerID),
		zap.Float64("cpu_usage", metric.CPUUsage),
		zap.Float64("memory_usage", metric.MemoryUsage),
	)

	return nil
}

func (s *AnalyticsServiceImpl) GetServerLoadAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	response, err := s.metricsRepo.GetServerLoadMetrics(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get server load analytics", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get server load analytics: %w", err)
	}

	return response, nil
}

func (s *AnalyticsServiceImpl) GetServerLoadStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	stats, err := s.metricsRepo.GetServerLoadStats(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get server load stats", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get server load stats: %w", err)
	}

	return stats, nil
}

func (s *AnalyticsServiceImpl) RecordError(ctx context.Context, metric domain.ErrorMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	if err := s.metricsRepo.SaveErrorMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save error metric",
			zap.String("error", err.Error()),
			zap.String("error_type", metric.ErrorType),
		)
		return fmt.Errorf("failed to save error metric: %w", err)
	}

	s.logger.Debug("Error metric recorded",
		zap.String("error_type", metric.ErrorType),
		zap.String("service", metric.Service),
	)

	return nil
}

func (s *AnalyticsServiceImpl) GetErrorAnalytics(ctx context.Context, opts domain.QueryOptions) (*domain.MetricResponse, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	response, err := s.metricsRepo.GetErrorMetrics(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get error analytics", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get error analytics: %w", err)
	}

	return response, nil
}

func (s *AnalyticsServiceImpl) GetErrorStats(ctx context.Context, opts domain.QueryOptions) (map[string]interface{}, error) {
	stats, err := s.metricsRepo.GetErrorStats(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to get error stats", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get error stats: %w", err)
	}

	return stats, nil
}

func (s *AnalyticsServiceImpl) GetTimeSeries(ctx context.Context, metricName string, opts domain.QueryOptions) ([]domain.Metric, error) {
	if opts.TimeRange.Start.IsZero() {
		opts.TimeRange.Start = time.Now().Add(-24 * time.Hour)
	}
	if opts.TimeRange.End.IsZero() {
		opts.TimeRange.End = time.Now()
	}

	if opts.Interval == "" {
		opts.Interval = "5m" // По умолчанию 5 минут
	}

	metrics, err := s.metricsRepo.GetTimeSeries(ctx, metricName, opts)
	if err != nil {
		s.logger.Error("Failed to get time series",
			zap.String("error", err.Error()),
			zap.String("metric_name", metricName),
		)
		return nil, fmt.Errorf("failed to get time series: %w", err)
	}

	return metrics, nil
}
