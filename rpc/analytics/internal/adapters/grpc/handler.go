package grpc

import (
	"context"
	"time"

	"github.com/par1ram/silence/rpc/analytics/api/proto"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AnalyticsHandler gRPC обработчик для analytics сервиса
type AnalyticsHandler struct {
	proto.UnimplementedAnalyticsServiceServer
	analyticsService ports.AnalyticsService
	logger           *zap.Logger
}

// NewAnalyticsHandler создает новый gRPC обработчик
func NewAnalyticsHandler(analyticsService ports.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// Health проверка здоровья сервиса
func (h *AnalyticsHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	h.logger.Debug("health check requested")

	return &proto.HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// CollectMetric собирает метрику
func (h *AnalyticsHandler) CollectMetric(ctx context.Context, req *proto.CollectMetricRequest) (*proto.CollectMetricResponse, error) {
	h.logger.Debug("collect metric requested", zap.String("name", req.Name))

	domainMetric := &domain.Metric{
		Name:      req.Name,
		Type:      req.Type,
		Value:     req.Value,
		Unit:      req.Unit,
		Tags:      req.Tags,
		Timestamp: time.Now(),
	}

	metric, err := h.analyticsService.CollectMetric(ctx, domainMetric)
	if err != nil {
		h.logger.Error("failed to collect metric", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to collect metric: %v", err)
	}

	return &proto.CollectMetricResponse{
		Success:  true,
		MetricId: metric.ID,
	}, nil
}

// GetMetrics получает метрики
func (h *AnalyticsHandler) GetMetrics(ctx context.Context, req *proto.GetMetricsRequest) (*proto.GetMetricsResponse, error) {
	h.logger.Debug("get metrics requested", zap.String("name", req.Name))

	filters := &domain.MetricFilters{
		Name:      req.Name,
		Tags:      req.Filters,
		Limit:     int(req.Limit),
		Offset:    int(req.Offset),
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
	}

	metrics, total, err := h.analyticsService.GetMetrics(ctx, filters)
	if err != nil {
		h.logger.Error("failed to get metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get metrics: %v", err)
	}

	protoMetrics := make([]*proto.Metric, len(metrics))
	for i, metric := range metrics {
		protoMetrics[i] = h.domainMetricToProto(metric)
	}

	return &proto.GetMetricsResponse{
		Metrics: protoMetrics,
		Total:   int32(total),
	}, nil
}

// GetMetricsHistory получает историю метрик
func (h *AnalyticsHandler) GetMetricsHistory(ctx context.Context, req *proto.GetMetricsHistoryRequest) (*proto.GetMetricsHistoryResponse, error) {
	h.logger.Debug("get metrics history requested", zap.String("name", req.Name))

	historyReq := &domain.MetricHistoryRequest{
		Name:      req.Name,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		Interval:  req.Interval,
	}

	points, err := h.analyticsService.GetMetricsHistory(ctx, historyReq)
	if err != nil {
		h.logger.Error("failed to get metrics history", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get metrics history: %v", err)
	}

	protoPoints := make([]*proto.TimeSeriesPoint, len(points))
	for i, point := range points {
		protoPoints[i] = &proto.TimeSeriesPoint{
			Timestamp: timestamppb.New(point.Timestamp),
			Value:     point.Value,
		}
	}

	return &proto.GetMetricsHistoryResponse{
		Points: protoPoints,
	}, nil
}

// GetStatistics получает статистику
func (h *AnalyticsHandler) GetStatistics(ctx context.Context, req *proto.GetStatisticsRequest) (*proto.GetStatisticsResponse, error) {
	h.logger.Debug("get statistics requested", zap.String("type", req.Type))

	statsReq := &domain.StatisticsRequest{
		Type:      req.Type,
		Period:    req.Period,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
	}

	stats, err := h.analyticsService.GetStatistics(ctx, statsReq)
	if err != nil {
		h.logger.Error("failed to get statistics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get statistics: %v", err)
	}

	protoStats := make([]*proto.Statistics, len(stats))
	for i, stat := range stats {
		protoStats[i] = &proto.Statistics{
			Id:           stat.ID,
			Name:         stat.Name,
			Type:         stat.Type,
			Value:        stat.Value,
			Unit:         stat.Unit,
			CalculatedAt: timestamppb.New(stat.CalculatedAt),
			Period:       stat.Period,
		}
	}

	return &proto.GetStatisticsResponse{
		Statistics: protoStats,
	}, nil
}

// GetSystemStats получает системную статистику
func (h *AnalyticsHandler) GetSystemStats(ctx context.Context, req *proto.GetSystemStatsRequest) (*proto.GetSystemStatsResponse, error) {
	h.logger.Debug("get system stats requested")

	stats, err := h.analyticsService.GetSystemStats(ctx)
	if err != nil {
		h.logger.Error("failed to get system stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get system stats: %v", err)
	}

	return &proto.GetSystemStatsResponse{
		Stats: &proto.SystemStats{
			TotalUsers:           stats.TotalUsers,
			ActiveUsers:          stats.ActiveUsers,
			TotalConnections:     stats.TotalConnections,
			ActiveConnections:    stats.ActiveConnections,
			TotalDataTransferred: stats.TotalDataTransferred,
			ServersCount:         stats.ServersCount,
			ActiveServers:        stats.ActiveServers,
			AvgConnectionTime:    stats.AvgConnectionTime,
			SystemLoad:           stats.SystemLoad,
			LastUpdated:          timestamppb.New(stats.LastUpdated),
		},
	}, nil
}

// GetUserStats получает статистику пользователя
func (h *AnalyticsHandler) GetUserStats(ctx context.Context, req *proto.GetUserStatsRequest) (*proto.GetUserStatsResponse, error) {
	h.logger.Debug("get user stats requested", zap.String("user_id", req.UserId))

	userStatsReq := &domain.UserStatsRequest{
		UserID:    req.UserId,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
	}

	stats, err := h.analyticsService.GetUserStats(ctx, userStatsReq)
	if err != nil {
		h.logger.Error("failed to get user stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user stats: %v", err)
	}

	return &proto.GetUserStatsResponse{
		Stats: &proto.UserStats{
			UserId:               stats.UserID,
			TotalConnections:     stats.TotalConnections,
			TotalDataTransferred: stats.TotalDataTransferred,
			TotalSessionTime:     stats.TotalSessionTime,
			FavoriteServersCount: int32(stats.FavoriteServersCount),
			AvgConnectionTime:    stats.AvgConnectionTime,
			FirstConnection:      timestamppb.New(stats.FirstConnection),
			LastConnection:       timestamppb.New(stats.LastConnection),
		},
	}, nil
}

// GetDashboardData получает данные для дашборда
func (h *AnalyticsHandler) GetDashboardData(ctx context.Context, req *proto.GetDashboardDataRequest) (*proto.GetDashboardDataResponse, error) {
	h.logger.Debug("get dashboard data requested", zap.String("time_range", req.TimeRange))

	data, err := h.analyticsService.GetDashboardData(ctx, req.TimeRange)
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get dashboard data: %v", err)
	}

	// Конвертируем connections over time
	connectionsOverTime := make([]*proto.TimeSeriesPoint, len(data.ConnectionsOverTime))
	for i, point := range data.ConnectionsOverTime {
		connectionsOverTime[i] = &proto.TimeSeriesPoint{
			Timestamp: timestamppb.New(point.Timestamp),
			Value:     point.Value,
		}
	}

	// Конвертируем data transfer over time
	dataTransferOverTime := make([]*proto.TimeSeriesPoint, len(data.DataTransferOverTime))
	for i, point := range data.DataTransferOverTime {
		dataTransferOverTime[i] = &proto.TimeSeriesPoint{
			Timestamp: timestamppb.New(point.Timestamp),
			Value:     point.Value,
		}
	}

	// Конвертируем server usage
	serverUsage := make([]*proto.ServerUsage, len(data.ServerUsage))
	for i, usage := range data.ServerUsage {
		serverUsage[i] = &proto.ServerUsage{
			ServerId:          usage.ServerID,
			ServerName:        usage.ServerName,
			ActiveConnections: usage.ActiveConnections,
			CpuUsage:          usage.CPUUsage,
			MemoryUsage:       usage.MemoryUsage,
			NetworkUsage:      usage.NetworkUsage,
		}
	}

	// Конвертируем region stats
	regionStats := make([]*proto.RegionStats, len(data.RegionStats))
	for i, stat := range data.RegionStats {
		regionStats[i] = &proto.RegionStats{
			Region:          stat.Region,
			UserCount:       stat.UserCount,
			ConnectionCount: stat.ConnectionCount,
			DataTransferred: stat.DataTransferred,
			AvgLatency:      stat.AvgLatency,
		}
	}

	// Конвертируем alerts
	alerts := make([]*proto.Alert, len(data.Alerts))
	for i, alert := range data.Alerts {
		alerts[i] = &proto.Alert{
			Id:           alert.ID,
			Type:         alert.RuleID,
			Severity:     string(alert.Severity),
			Title:        alert.Message,
			Message:      alert.Message,
			CreatedAt:    timestamppb.New(alert.CreatedAt),
			Acknowledged: alert.Status == "acknowledged",
		}
	}

	return &proto.GetDashboardDataResponse{
		Data: &proto.DashboardData{
			SystemStats: &proto.SystemStats{
				TotalUsers:           data.SystemStats.TotalUsers,
				ActiveUsers:          data.SystemStats.ActiveUsers,
				TotalConnections:     data.SystemStats.TotalConnections,
				ActiveConnections:    data.SystemStats.ActiveConnections,
				TotalDataTransferred: data.SystemStats.TotalDataTransferred,
				ServersCount:         data.SystemStats.ServersCount,
				ActiveServers:        data.SystemStats.ActiveServers,
				AvgConnectionTime:    data.SystemStats.AvgConnectionTime,
				SystemLoad:           data.SystemStats.SystemLoad,
				LastUpdated:          timestamppb.New(data.SystemStats.LastUpdated),
			},
			ConnectionsOverTime:  connectionsOverTime,
			DataTransferOverTime: dataTransferOverTime,
			ServerUsage:          serverUsage,
			RegionStats:          regionStats,
			Alerts:               alerts,
		},
	}, nil
}

// PredictLoad предсказывает нагрузку
func (h *AnalyticsHandler) PredictLoad(ctx context.Context, req *proto.PredictLoadRequest) (*proto.PredictLoadResponse, error) {
	h.logger.Debug("predict load requested", zap.String("server_id", req.ServerId))

	predictionReq := &domain.PredictionRequest{
		ServerID:   req.ServerId,
		HoursAhead: int(req.HoursAhead),
	}

	predictions, err := h.analyticsService.PredictLoad(ctx, predictionReq)
	if err != nil {
		h.logger.Error("failed to predict load", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to predict load: %v", err)
	}

	protoPredictions := make([]*proto.PredictionPoint, len(predictions))
	for i, prediction := range predictions {
		protoPredictions[i] = &proto.PredictionPoint{
			Timestamp:      timestamppb.New(prediction.Timestamp),
			PredictedValue: prediction.PredictedValue,
			Confidence:     prediction.Confidence,
		}
	}

	return &proto.PredictLoadResponse{
		Predictions: protoPredictions,
	}, nil
}

// PredictTrend предсказывает тренд
func (h *AnalyticsHandler) PredictTrend(ctx context.Context, req *proto.PredictTrendRequest) (*proto.PredictTrendResponse, error) {
	h.logger.Debug("predict trend requested", zap.String("metric_name", req.MetricName))

	trendReq := &domain.TrendRequest{
		MetricName: req.MetricName,
		DaysAhead:  int(req.DaysAhead),
	}

	predictions, err := h.analyticsService.PredictTrend(ctx, trendReq)
	if err != nil {
		h.logger.Error("failed to predict trend", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to predict trend: %v", err)
	}

	protoPredictions := make([]*proto.PredictionPoint, len(predictions))
	for i, prediction := range predictions {
		protoPredictions[i] = &proto.PredictionPoint{
			Timestamp:      timestamppb.New(prediction.Timestamp),
			PredictedValue: prediction.PredictedValue,
			Confidence:     prediction.Confidence,
		}
	}

	return &proto.PredictTrendResponse{
		Predictions: protoPredictions,
	}, nil
}

// domainMetricToProto converts domain metric to proto metric
func (h *AnalyticsHandler) domainMetricToProto(metric *domain.Metric) *proto.Metric {
	return &proto.Metric{
		Id:        metric.ID,
		Name:      metric.Name,
		Type:      metric.Type,
		Value:     metric.Value,
		Unit:      metric.Unit,
		Tags:      metric.Tags,
		Timestamp: timestamppb.New(metric.Timestamp),
	}
}
