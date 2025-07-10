package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TelemetryManager управляет OpenTelemetry компонентами
type TelemetryManager struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *log.LoggerProvider
	tracer         trace.Tracer
	meter          metric.Meter
	logger         *zap.Logger
	config         config.OTelConfig
	shutdownFuncs  []func(context.Context) error
}

// NewTelemetryManager создает новый менеджер телеметрии
func NewTelemetryManager(cfg config.OTelConfig, logger *zap.Logger) (*TelemetryManager, error) {
	tm := &TelemetryManager{
		config:        cfg,
		logger:        logger,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Создаем ресурс
	res, err := tm.createResource()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Инициализируем трейсинг
	if cfg.TracingEnabled {
		if err := tm.initTracing(res); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Инициализируем метрики
	if cfg.MetricsEnabled {
		if err := tm.initMetrics(res); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Инициализируем логирование
	if cfg.LoggingEnabled {
		if err := tm.initLogging(res); err != nil {
			return nil, fmt.Errorf("failed to initialize logging: %w", err)
		}
	}

	// Настраиваем пропагацию
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tm, nil
}

// createResource создает ресурс OpenTelemetry
func (tm *TelemetryManager) createResource() (*resource.Resource, error) {
	attrs := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithProcess(),
	}

	// Добавляем кастомные атрибуты
	if len(tm.config.ResourceAttributes) > 0 {
		resourceAttrs := make([]attribute.KeyValue, 0, len(tm.config.ResourceAttributes))
		for k, v := range tm.config.ResourceAttributes {
			resourceAttrs = append(resourceAttrs, attribute.String(k, v))
		}
		attrs = append(attrs, resource.WithAttributes(resourceAttrs...))
	}

	return resource.New(context.Background(), attrs...)
}

// initTracing инициализирует трейсинг
func (tm *TelemetryManager) initTracing(res *resource.Resource) error {
	var exporters []sdktrace.SpanExporter

	// Jaeger экспортер
	if tm.config.JaegerEndpoint != "" {
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(tm.config.JaegerEndpoint),
		))
		if err != nil {
			tm.logger.Warn("Failed to create Jaeger exporter", zap.Error(err))
		} else {
			exporters = append(exporters, jaegerExporter)
		}
	}

	// Zipkin экспортер
	if tm.config.ZipkinEndpoint != "" {
		zipkinExporter, err := zipkin.New(tm.config.ZipkinEndpoint)
		if err != nil {
			tm.logger.Warn("Failed to create Zipkin exporter", zap.Error(err))
		} else {
			exporters = append(exporters, zipkinExporter)
		}
	}

	// OTLP gRPC экспортер
	if tm.config.OTLPTraceEndpoint != "" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(tm.config.OTLPTraceEndpoint),
		}
		if tm.config.OTLPTraceInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		otlpExporter, err := otlptracegrpc.New(context.Background(), opts...)
		if err != nil {
			tm.logger.Warn("Failed to create OTLP gRPC trace exporter", zap.Error(err))
		} else {
			exporters = append(exporters, otlpExporter)
		}
	}

	// Fallback к stdout если нет других экспортеров
	if len(exporters) == 0 {
		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
		exporters = append(exporters, stdoutExporter)
	}

	// Создаем span processor
	var processors []sdktrace.SpanProcessor
	for _, exporter := range exporters {
		processors = append(processors, sdktrace.NewBatchSpanProcessor(exporter))
	}

	// Создаем tracer provider
	tm.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(tm.config.TraceSamplingRatio))),
	)

	// Добавляем processors
	for _, processor := range processors {
		tm.tracerProvider.RegisterSpanProcessor(processor)
	}

	// Устанавливаем глобальный tracer provider
	otel.SetTracerProvider(tm.tracerProvider)

	// Создаем tracer
	tm.tracer = tm.tracerProvider.Tracer(
		tm.config.ServiceName,
		trace.WithInstrumentationVersion(tm.config.ServiceVersion),
		trace.WithSchemaURL("https://opentelemetry.io/schemas/1.21.0"),
	)

	// Добавляем функцию shutdown
	tm.shutdownFuncs = append(tm.shutdownFuncs, func(ctx context.Context) error {
		return tm.tracerProvider.Shutdown(ctx)
	})

	return nil
}

// initMetrics инициализирует метрики
func (tm *TelemetryManager) initMetrics(res *resource.Resource) error {
	var readers []sdkmetric.Reader

	// Prometheus экспортер
	if tm.config.PrometheusEndpoint != "" {
		promExporter, err := prometheus.New()
		if err != nil {
			tm.logger.Warn("Failed to create Prometheus exporter", zap.Error(err))
		} else {
			readers = append(readers, promExporter)

			// Запускаем HTTP сервер для метрик
			go func() {
				mux := http.NewServeMux()
				mux.Handle("/metrics", promhttp.Handler())
				server := &http.Server{
					Addr:    ":" + tm.config.MetricsPort,
					Handler: mux,
				}
				tm.logger.Info("Starting Prometheus metrics server", zap.String("addr", server.Addr))
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					tm.logger.Error("Prometheus metrics server error", zap.Error(err))
				}
			}()
		}
	}

	// OTLP gRPC метрики экспортер
	if tm.config.OTLPMetricEndpoint != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctx, tm.config.OTLPMetricEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			tm.logger.Warn("Failed to connect to OTLP gRPC endpoint for metrics", zap.Error(err))
		} else {
			otlpExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
			if err != nil {
				tm.logger.Warn("Failed to create OTLP gRPC metrics exporter", zap.Error(err))
			} else {
				readers = append(readers, sdkmetric.NewPeriodicReader(otlpExporter, sdkmetric.WithInterval(15*time.Second)))
			}
		}
	}

	// Fallback к stdout если нет других экспортеров
	if len(readers) == 0 {
		stdoutExporter, err := stdoutmetric.New()
		if err != nil {
			return fmt.Errorf("failed to create stdout metrics exporter: %w", err)
		}
		readers = append(readers, sdkmetric.NewPeriodicReader(stdoutExporter, sdkmetric.WithInterval(30*time.Second)))
	}

	// Создаем meter provider
	var opts []sdkmetric.Option
	opts = append(opts, sdkmetric.WithResource(res))
	for _, reader := range readers {
		opts = append(opts, sdkmetric.WithReader(reader))
	}
	tm.meterProvider = sdkmetric.NewMeterProvider(opts...)

	// Устанавливаем глобальный meter provider
	otel.SetMeterProvider(tm.meterProvider)

	// Создаем meter
	tm.meter = tm.meterProvider.Meter(
		tm.config.ServiceName,
		metric.WithInstrumentationVersion(tm.config.ServiceVersion),
		metric.WithSchemaURL("https://opentelemetry.io/schemas/1.21.0"),
	)

	// Добавляем функцию shutdown
	tm.shutdownFuncs = append(tm.shutdownFuncs, func(ctx context.Context) error {
		return tm.meterProvider.Shutdown(ctx)
	})

	return nil
}

// initLogging инициализирует логирование
func (tm *TelemetryManager) initLogging(res *resource.Resource) error {
	var exporters []log.Exporter

	// OTLP HTTP лог экспортер
	if tm.config.OTLPLogEndpoint != "" {
		// Пока используем stdout, так как OTLP log экспортер еще в разработке
		tm.logger.Info("OTLP log exporter not fully implemented yet, using stdout")
	}

	// Stdout экспортер
	stdoutExporter, err := stdoutlog.New()
	if err != nil {
		return fmt.Errorf("failed to create stdout log exporter: %w", err)
	}
	exporters = append(exporters, stdoutExporter)

	// Создаем log provider
	tm.loggerProvider = log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exporters[0])),
	)

	// Устанавливаем глобальный logger provider
	global.SetLoggerProvider(tm.loggerProvider)

	// Добавляем функцию shutdown
	tm.shutdownFuncs = append(tm.shutdownFuncs, func(ctx context.Context) error {
		return tm.loggerProvider.Shutdown(ctx)
	})

	return nil
}

// GetTracer возвращает tracer
func (tm *TelemetryManager) GetTracer() trace.Tracer {
	return tm.tracer
}

// GetMeter возвращает meter
func (tm *TelemetryManager) GetMeter() metric.Meter {
	return tm.meter
}

// GetTracerProvider возвращает tracer provider
func (tm *TelemetryManager) GetTracerProvider() *sdktrace.TracerProvider {
	return tm.tracerProvider
}

// GetMeterProvider возвращает meter provider
func (tm *TelemetryManager) GetMeterProvider() *sdkmetric.MeterProvider {
	return tm.meterProvider
}

// Shutdown останавливает все компоненты телеметрии
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	var errors []error

	for _, shutdown := range tm.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}
