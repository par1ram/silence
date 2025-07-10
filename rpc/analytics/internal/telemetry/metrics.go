package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// MetricsCollector собирает метрики для аналитики
type MetricsCollector struct {
	meter  metric.Meter
	logger *zap.Logger

	// Connection metrics
	activeConnections  metric.Int64UpDownCounter
	totalConnections   metric.Int64Counter
	connectionDuration metric.Float64Histogram
	connectionDataIn   metric.Int64Counter
	connectionDataOut  metric.Int64Counter
	connectionErrors   metric.Int64Counter

	// Server metrics
	serverCPUUsage    metric.Float64Gauge
	serverMemoryUsage metric.Float64Gauge
	serverNetworkIn   metric.Float64Counter
	serverNetworkOut  metric.Float64Counter
	serverConnections metric.Int64UpDownCounter
	serverLoad        metric.Float64Gauge

	// User metrics
	activeUsers         metric.Int64UpDownCounter
	totalUsers          metric.Int64Counter
	userSessions        metric.Int64Counter
	userDataTransferred metric.Int64Counter
	userActivityTime    metric.Float64Histogram

	// Bypass metrics
	bypassAttempts   metric.Int64Counter
	bypassSuccess    metric.Int64Counter
	bypassLatency    metric.Float64Histogram
	bypassThroughput metric.Float64Histogram
	bypassBlocked    metric.Int64Counter

	// System metrics
	systemLoad         metric.Float64Gauge
	systemUptime       metric.Float64Counter
	systemErrors       metric.Int64Counter
	systemRequests     metric.Int64Counter
	systemResponseTime metric.Float64Histogram

	// Analytics metrics
	metricsProcessed     metric.Int64Counter
	metricsErrors        metric.Int64Counter
	metricsLatency       metric.Float64Histogram
	dashboardRequests    metric.Int64Counter
	predictionsGenerated metric.Int64Counter
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector(meter metric.Meter, logger *zap.Logger) (*MetricsCollector, error) {
	mc := &MetricsCollector{
		meter:  meter,
		logger: logger,
	}

	if err := mc.initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return mc, nil
}

// initMetrics инициализирует все метрики
func (mc *MetricsCollector) initMetrics() error {
	var err error

	// Connection metrics
	mc.activeConnections, err = mc.meter.Int64UpDownCounter(
		"silence_active_connections",
		metric.WithDescription("Number of active VPN connections"),
		metric.WithUnit("connections"),
	)
	if err != nil {
		return fmt.Errorf("failed to create active_connections metric: %w", err)
	}

	mc.totalConnections, err = mc.meter.Int64Counter(
		"silence_total_connections",
		metric.WithDescription("Total number of VPN connections"),
		metric.WithUnit("connections"),
	)
	if err != nil {
		return fmt.Errorf("failed to create total_connections metric: %w", err)
	}

	mc.connectionDuration, err = mc.meter.Float64Histogram(
		"silence_connection_duration_seconds",
		metric.WithDescription("Duration of VPN connections"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create connection_duration metric: %w", err)
	}

	mc.connectionDataIn, err = mc.meter.Int64Counter(
		"silence_connection_data_in_bytes",
		metric.WithDescription("Incoming data through VPN connections"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("failed to create connection_data_in metric: %w", err)
	}

	mc.connectionDataOut, err = mc.meter.Int64Counter(
		"silence_connection_data_out_bytes",
		metric.WithDescription("Outgoing data through VPN connections"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("failed to create connection_data_out metric: %w", err)
	}

	mc.connectionErrors, err = mc.meter.Int64Counter(
		"silence_connection_errors",
		metric.WithDescription("Number of connection errors"),
		metric.WithUnit("errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create connection_errors metric: %w", err)
	}

	// Server metrics
	mc.serverCPUUsage, err = mc.meter.Float64Gauge(
		"silence_server_cpu_usage_percent",
		metric.WithDescription("Server CPU usage percentage"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_cpu_usage metric: %w", err)
	}

	mc.serverMemoryUsage, err = mc.meter.Float64Gauge(
		"silence_server_memory_usage_percent",
		metric.WithDescription("Server memory usage percentage"),
		metric.WithUnit("percent"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_memory_usage metric: %w", err)
	}

	mc.serverNetworkIn, err = mc.meter.Float64Counter(
		"silence_server_network_in_bytes",
		metric.WithDescription("Server incoming network traffic"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_network_in metric: %w", err)
	}

	mc.serverNetworkOut, err = mc.meter.Float64Counter(
		"silence_server_network_out_bytes",
		metric.WithDescription("Server outgoing network traffic"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_network_out metric: %w", err)
	}

	mc.serverConnections, err = mc.meter.Int64UpDownCounter(
		"silence_server_connections",
		metric.WithDescription("Number of connections per server"),
		metric.WithUnit("connections"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_connections metric: %w", err)
	}

	mc.serverLoad, err = mc.meter.Float64Gauge(
		"silence_server_load",
		metric.WithDescription("Server load average"),
		metric.WithUnit("load"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server_load metric: %w", err)
	}

	// User metrics
	mc.activeUsers, err = mc.meter.Int64UpDownCounter(
		"silence_active_users",
		metric.WithDescription("Number of active users"),
		metric.WithUnit("users"),
	)
	if err != nil {
		return fmt.Errorf("failed to create active_users metric: %w", err)
	}

	mc.totalUsers, err = mc.meter.Int64Counter(
		"silence_total_users",
		metric.WithDescription("Total number of users"),
		metric.WithUnit("users"),
	)
	if err != nil {
		return fmt.Errorf("failed to create total_users metric: %w", err)
	}

	mc.userSessions, err = mc.meter.Int64Counter(
		"silence_user_sessions",
		metric.WithDescription("Number of user sessions"),
		metric.WithUnit("sessions"),
	)
	if err != nil {
		return fmt.Errorf("failed to create user_sessions metric: %w", err)
	}

	mc.userDataTransferred, err = mc.meter.Int64Counter(
		"silence_user_data_transferred_bytes",
		metric.WithDescription("Total data transferred by users"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return fmt.Errorf("failed to create user_data_transferred metric: %w", err)
	}

	mc.userActivityTime, err = mc.meter.Float64Histogram(
		"silence_user_activity_time_seconds",
		metric.WithDescription("User activity time"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create user_activity_time metric: %w", err)
	}

	// Bypass metrics
	mc.bypassAttempts, err = mc.meter.Int64Counter(
		"silence_bypass_attempts",
		metric.WithDescription("Number of DPI bypass attempts"),
		metric.WithUnit("attempts"),
	)
	if err != nil {
		return fmt.Errorf("failed to create bypass_attempts metric: %w", err)
	}

	mc.bypassSuccess, err = mc.meter.Int64Counter(
		"silence_bypass_success",
		metric.WithDescription("Number of successful DPI bypass attempts"),
		metric.WithUnit("attempts"),
	)
	if err != nil {
		return fmt.Errorf("failed to create bypass_success metric: %w", err)
	}

	mc.bypassLatency, err = mc.meter.Float64Histogram(
		"silence_bypass_latency_seconds",
		metric.WithDescription("DPI bypass latency"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create bypass_latency metric: %w", err)
	}

	mc.bypassThroughput, err = mc.meter.Float64Histogram(
		"silence_bypass_throughput_mbps",
		metric.WithDescription("DPI bypass throughput"),
		metric.WithUnit("mbps"),
	)
	if err != nil {
		return fmt.Errorf("failed to create bypass_throughput metric: %w", err)
	}

	mc.bypassBlocked, err = mc.meter.Int64Counter(
		"silence_bypass_blocked",
		metric.WithDescription("Number of blocked bypass attempts"),
		metric.WithUnit("attempts"),
	)
	if err != nil {
		return fmt.Errorf("failed to create bypass_blocked metric: %w", err)
	}

	// System metrics
	mc.systemLoad, err = mc.meter.Float64Gauge(
		"silence_system_load",
		metric.WithDescription("System load average"),
		metric.WithUnit("load"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_load metric: %w", err)
	}

	mc.systemUptime, err = mc.meter.Float64Counter(
		"silence_system_uptime_seconds",
		metric.WithDescription("System uptime"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_uptime metric: %w", err)
	}

	mc.systemErrors, err = mc.meter.Int64Counter(
		"silence_system_errors",
		metric.WithDescription("Number of system errors"),
		metric.WithUnit("errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_errors metric: %w", err)
	}

	mc.systemRequests, err = mc.meter.Int64Counter(
		"silence_system_requests",
		metric.WithDescription("Number of system requests"),
		metric.WithUnit("requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_requests metric: %w", err)
	}

	mc.systemResponseTime, err = mc.meter.Float64Histogram(
		"silence_system_response_time_seconds",
		metric.WithDescription("System response time"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_response_time metric: %w", err)
	}

	// Analytics metrics
	mc.metricsProcessed, err = mc.meter.Int64Counter(
		"silence_metrics_processed",
		metric.WithDescription("Number of processed metrics"),
		metric.WithUnit("metrics"),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics_processed metric: %w", err)
	}

	mc.metricsErrors, err = mc.meter.Int64Counter(
		"silence_metrics_errors",
		metric.WithDescription("Number of metrics processing errors"),
		metric.WithUnit("errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics_errors metric: %w", err)
	}

	mc.metricsLatency, err = mc.meter.Float64Histogram(
		"silence_metrics_latency_seconds",
		metric.WithDescription("Metrics processing latency"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics_latency metric: %w", err)
	}

	mc.dashboardRequests, err = mc.meter.Int64Counter(
		"silence_dashboard_requests",
		metric.WithDescription("Number of dashboard requests"),
		metric.WithUnit("requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create dashboard_requests metric: %w", err)
	}

	mc.predictionsGenerated, err = mc.meter.Int64Counter(
		"silence_predictions_generated",
		metric.WithDescription("Number of predictions generated"),
		metric.WithUnit("predictions"),
	)
	if err != nil {
		return fmt.Errorf("failed to create predictions_generated metric: %w", err)
	}

	return nil
}

// Connection metrics methods
func (mc *MetricsCollector) RecordActiveConnections(ctx context.Context, count int64, serverID, region string) {
	mc.activeConnections.Add(ctx, count, metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordNewConnection(ctx context.Context, userID, serverID, protocol, region string) {
	mc.totalConnections.Add(ctx, 1, metric.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("server_id", serverID),
		attribute.String("protocol", protocol),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordConnectionDuration(ctx context.Context, duration time.Duration, serverID, region string) {
	mc.connectionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordConnectionData(ctx context.Context, bytesIn, bytesOut int64, serverID, region string) {
	attrs := metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	)
	mc.connectionDataIn.Add(ctx, bytesIn, attrs)
	mc.connectionDataOut.Add(ctx, bytesOut, attrs)
}

func (mc *MetricsCollector) RecordConnectionError(ctx context.Context, errorType, serverID, region string) {
	mc.connectionErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

// Server metrics methods
func (mc *MetricsCollector) RecordServerCPUUsage(ctx context.Context, usage float64, serverID, region string) {
	mc.serverCPUUsage.Record(ctx, usage, metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordServerMemoryUsage(ctx context.Context, usage float64, serverID, region string) {
	mc.serverMemoryUsage.Record(ctx, usage, metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordServerNetwork(ctx context.Context, bytesIn, bytesOut float64, serverID, region string) {
	attrs := metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	)
	mc.serverNetworkIn.Add(ctx, bytesIn, attrs)
	mc.serverNetworkOut.Add(ctx, bytesOut, attrs)
}

func (mc *MetricsCollector) RecordServerConnections(ctx context.Context, count int64, serverID, region string) {
	mc.serverConnections.Add(ctx, count, metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordServerLoad(ctx context.Context, load float64, serverID, region string) {
	mc.serverLoad.Record(ctx, load, metric.WithAttributes(
		attribute.String("server_id", serverID),
		attribute.String("region", region),
	))
}

// User metrics methods
func (mc *MetricsCollector) RecordActiveUsers(ctx context.Context, count int64, region string) {
	mc.activeUsers.Add(ctx, count, metric.WithAttributes(
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordNewUser(ctx context.Context, region string) {
	mc.totalUsers.Add(ctx, 1, metric.WithAttributes(
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordUserSession(ctx context.Context, userID, region string) {
	mc.userSessions.Add(ctx, 1, metric.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordUserDataTransferred(ctx context.Context, bytes int64, userID, region string) {
	mc.userDataTransferred.Add(ctx, bytes, metric.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordUserActivityTime(ctx context.Context, duration time.Duration, userID, region string) {
	mc.userActivityTime.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("region", region),
	))
}

// Bypass metrics methods
func (mc *MetricsCollector) RecordBypassAttempt(ctx context.Context, bypassType, region string) {
	mc.bypassAttempts.Add(ctx, 1, metric.WithAttributes(
		attribute.String("bypass_type", bypassType),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordBypassSuccess(ctx context.Context, bypassType, region string) {
	mc.bypassSuccess.Add(ctx, 1, metric.WithAttributes(
		attribute.String("bypass_type", bypassType),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordBypassLatency(ctx context.Context, latency time.Duration, bypassType, region string) {
	mc.bypassLatency.Record(ctx, latency.Seconds(), metric.WithAttributes(
		attribute.String("bypass_type", bypassType),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordBypassThroughput(ctx context.Context, throughput float64, bypassType, region string) {
	mc.bypassThroughput.Record(ctx, throughput, metric.WithAttributes(
		attribute.String("bypass_type", bypassType),
		attribute.String("region", region),
	))
}

func (mc *MetricsCollector) RecordBypassBlocked(ctx context.Context, bypassType, region string) {
	mc.bypassBlocked.Add(ctx, 1, metric.WithAttributes(
		attribute.String("bypass_type", bypassType),
		attribute.String("region", region),
	))
}

// System metrics methods
func (mc *MetricsCollector) RecordSystemLoad(ctx context.Context, load float64) {
	mc.systemLoad.Record(ctx, load)
}

func (mc *MetricsCollector) RecordSystemUptime(ctx context.Context, uptime time.Duration) {
	mc.systemUptime.Add(ctx, uptime.Seconds())
}

func (mc *MetricsCollector) RecordSystemError(ctx context.Context, errorType, service string) {
	mc.systemErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
		attribute.String("service", service),
	))
}

func (mc *MetricsCollector) RecordSystemRequest(ctx context.Context, method, endpoint string) {
	mc.systemRequests.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	))
}

func (mc *MetricsCollector) RecordSystemResponseTime(ctx context.Context, duration time.Duration, method, endpoint string) {
	mc.systemResponseTime.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	))
}

// Analytics metrics methods
func (mc *MetricsCollector) RecordMetricProcessed(ctx context.Context, metricType string) {
	mc.metricsProcessed.Add(ctx, 1, metric.WithAttributes(
		attribute.String("metric_type", metricType),
	))
}

func (mc *MetricsCollector) RecordMetricError(ctx context.Context, errorType, metricType string) {
	mc.metricsErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
		attribute.String("metric_type", metricType),
	))
}

func (mc *MetricsCollector) RecordMetricLatency(ctx context.Context, duration time.Duration, metricType string) {
	mc.metricsLatency.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("metric_type", metricType),
	))
}

func (mc *MetricsCollector) RecordDashboardRequest(ctx context.Context, dashboardType string) {
	mc.dashboardRequests.Add(ctx, 1, metric.WithAttributes(
		attribute.String("dashboard_type", dashboardType),
	))
}

func (mc *MetricsCollector) RecordPredictionGenerated(ctx context.Context, predictionType string) {
	mc.predictionsGenerated.Add(ctx, 1, metric.WithAttributes(
		attribute.String("prediction_type", predictionType),
	))
}
