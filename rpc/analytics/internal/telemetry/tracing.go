package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingManager управляет трейсингом для аналитики
type TracingManager struct {
	tracer trace.Tracer
	logger *zap.Logger
}

// NewTracingManager создает новый менеджер трейсинга
func NewTracingManager(tracer trace.Tracer, logger *zap.Logger) *TracingManager {
	return &TracingManager{
		tracer: tracer,
		logger: logger,
	}
}

// StartSpan начинает новый span
func (tm *TracingManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tm.tracer.Start(ctx, name, opts...)
}

// TraceAnalyticsOperation трейсит операцию аналитики
func (tm *TracingManager) TraceAnalyticsOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("analytics.%s", operation))
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "analytics"),
		attribute.String("operation.name", operation),
		attribute.String("service.name", "analytics"),
	)

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		tm.logger.Error("Analytics operation failed",
			zap.String("operation", operation),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
	} else {
		span.SetStatus(codes.Ok, "")
		tm.logger.Debug("Analytics operation completed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// TraceMetricCollection трейсит сбор метрик
func (tm *TracingManager) TraceMetricCollection(ctx context.Context, metricName string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "metric.collect")
	defer span.End()

	span.SetAttributes(
		attribute.String("metric.name", metricName),
		attribute.String("operation.type", "metric_collection"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceMetricQuery трейсит запрос метрик
func (tm *TracingManager) TraceMetricQuery(ctx context.Context, query string, filters map[string]string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "metric.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.String("operation.type", "metric_query"),
	)

	// Добавляем фильтры как атрибуты
	for key, value := range filters {
		span.SetAttributes(attribute.String(fmt.Sprintf("filter.%s", key), value))
	}

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceDashboardRequest трейсит запрос дашборда
func (tm *TracingManager) TraceDashboardRequest(ctx context.Context, dashboardID string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "dashboard.request")
	defer span.End()

	span.SetAttributes(
		attribute.String("dashboard.id", dashboardID),
		attribute.String("operation.type", "dashboard_request"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TracePrediction трейсит создание предсказания
func (tm *TracingManager) TracePrediction(ctx context.Context, predictionType string, hoursAhead int, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "prediction.generate")
	defer span.End()

	span.SetAttributes(
		attribute.String("prediction.type", predictionType),
		attribute.Int("prediction.hours_ahead", hoursAhead),
		attribute.String("operation.type", "prediction"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceUserActivity трейсит активность пользователя
func (tm *TracingManager) TraceUserActivity(ctx context.Context, userID string, activityType string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "user.activity")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("activity.type", activityType),
		attribute.String("operation.type", "user_activity"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceServerLoad трейсит нагрузку сервера
func (tm *TracingManager) TraceServerLoad(ctx context.Context, serverID string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "server.load")
	defer span.End()

	span.SetAttributes(
		attribute.String("server.id", serverID),
		attribute.String("operation.type", "server_load"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceConnectionMetrics трейсит метрики подключений
func (tm *TracingManager) TraceConnectionMetrics(ctx context.Context, connectionID string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "connection.metrics")
	defer span.End()

	span.SetAttributes(
		attribute.String("connection.id", connectionID),
		attribute.String("operation.type", "connection_metrics"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceBypassEffectiveness трейсит эффективность обхода
func (tm *TracingManager) TraceBypassEffectiveness(ctx context.Context, bypassType string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "bypass.effectiveness")
	defer span.End()

	span.SetAttributes(
		attribute.String("bypass.type", bypassType),
		attribute.String("operation.type", "bypass_effectiveness"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceGRPCRequest трейсит gRPC запрос
func (tm *TracingManager) TraceGRPCRequest(ctx context.Context, method string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("grpc.%s", method))
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.method", method),
		attribute.String("rpc.service", "analytics"),
		attribute.String("rpc.system", "grpc"),
		attribute.String("operation.type", "grpc_request"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceDataProcessing трейсит обработку данных
func (tm *TracingManager) TraceDataProcessing(ctx context.Context, dataType string, recordCount int, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "data.processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("data.type", dataType),
		attribute.Int("data.record_count", recordCount),
		attribute.String("operation.type", "data_processing"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceAlert трейсит обработку алертов
func (tm *TracingManager) TraceAlert(ctx context.Context, alertType string, severity string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "alert.process")
	defer span.End()

	span.SetAttributes(
		attribute.String("alert.type", alertType),
		attribute.String("alert.severity", severity),
		attribute.String("operation.type", "alert_processing"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceRedisOperation трейсит операции с Redis
func (tm *TracingManager) TraceRedisOperation(ctx context.Context, operation string, key string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("redis.%s", operation))
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", operation),
		attribute.String("db.key", key),
		attribute.String("operation.type", "redis_operation"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceHTTPRequest трейсит HTTP запрос
func (tm *TracingManager) TraceHTTPRequest(ctx context.Context, method string, endpoint string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, fmt.Sprintf("http.%s", method))
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.endpoint", endpoint),
		attribute.String("operation.type", "http_request"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceDatabaseQuery трейсит запрос к базе данных
func (tm *TracingManager) TraceDatabaseQuery(ctx context.Context, query string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "db.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", query),
		attribute.String("operation.type", "db_query"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceSystemMetrics трейсит системные метрики
func (tm *TracingManager) TraceSystemMetrics(ctx context.Context, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "system.metrics")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "system_metrics"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceAggregation трейсит агрегацию данных
func (tm *TracingManager) TraceAggregation(ctx context.Context, aggregationType string, timeRange string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "data.aggregation")
	defer span.End()

	span.SetAttributes(
		attribute.String("aggregation.type", aggregationType),
		attribute.String("aggregation.time_range", timeRange),
		attribute.String("operation.type", "data_aggregation"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// TraceExport трейсит экспорт данных
func (tm *TracingManager) TraceExport(ctx context.Context, exportType string, format string, fn func(context.Context) error) error {
	ctx, span := tm.StartSpan(ctx, "data.export")
	defer span.End()

	span.SetAttributes(
		attribute.String("export.type", exportType),
		attribute.String("export.format", format),
		attribute.String("operation.type", "data_export"),
	)

	return tm.traceWithErrorHandling(ctx, span, fn)
}

// AddSpanEvent добавляет событие в активный span
func (tm *TracingManager) AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attributes...))
	}
}

// RecordError записывает ошибку в активный span
func (tm *TracingManager) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttribute устанавливает атрибут для активного span
func (tm *TracingManager) SetSpanAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		var attr attribute.KeyValue
		switch v := value.(type) {
		case string:
			attr = attribute.String(key, v)
		case int:
			attr = attribute.Int(key, v)
		case int64:
			attr = attribute.Int64(key, v)
		case float64:
			attr = attribute.Float64(key, v)
		case bool:
			attr = attribute.Bool(key, v)
		default:
			attr = attribute.String(key, fmt.Sprintf("%v", v))
		}
		span.SetAttributes(attr)
	}
}

// GetTraceID возвращает ID текущего трейса
func (tm *TracingManager) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID возвращает ID текущего span
func (tm *TracingManager) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// traceWithErrorHandling выполняет функцию с обработкой ошибок для трейсинга
func (tm *TracingManager) traceWithErrorHandling(ctx context.Context, span trace.Span, fn func(context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		tm.logger.Error("Traced operation failed",
			zap.String("span_name", span.SpanContext().TraceID().String()),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
	} else {
		span.SetStatus(codes.Ok, "")
		tm.logger.Debug("Traced operation completed",
			zap.String("span_name", span.SpanContext().TraceID().String()),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// CreateChildSpan создает дочерний span
func (tm *TracingManager) CreateChildSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attributes...),
	}
	return tm.StartSpan(ctx, name, opts...)
}

// StartTransaction начинает новую транзакцию (корневой span)
func (tm *TracingManager) StartTransaction(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attributes...),
	}
	return tm.StartSpan(ctx, name, opts...)
}

// FinishTransaction завершает транзакцию с результатом
func (tm *TracingManager) FinishTransaction(span trace.Span, err error, attributes ...attribute.KeyValue) {
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}
