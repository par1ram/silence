package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/adapters"
	"github.com/par1ram/silence/rpc/analytics/internal/config"

	"github.com/par1ram/silence/rpc/analytics/internal/services"
	"github.com/par1ram/silence/rpc/analytics/internal/telemetry"
	"github.com/par1ram/silence/shared/logger"
	"github.com/par1ram/silence/shared/redis"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	zapLogger := logger.NewLogger("analytics")
	defer zapLogger.Sync()

	zapLogger.Info("Starting Silence Analytics Service",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.OpenTelemetry.Environment),
	)

	// Инициализируем OpenTelemetry
	telemetryManager, err := telemetry.NewTelemetryManager(cfg.OpenTelemetry, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize OpenTelemetry", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetryManager.Shutdown(ctx); err != nil {
			zapLogger.Error("Failed to shutdown OpenTelemetry", zap.Error(err))
		}
	}()

	// Создаем метрики коллектор
	metricsCollector, err := telemetry.NewMetricsCollector(
		telemetryManager.GetMeter(),
		zapLogger,
	)
	if err != nil {
		zapLogger.Fatal("Failed to create metrics collector", zap.Error(err))
	}

	// Создаем трейсинг менеджер
	tracingManager := telemetry.NewTracingManager(
		telemetryManager.GetTracer(),
		zapLogger,
	)

	// Инициализируем Redis клиент
	redisClient, err := redis.NewClient(&redis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize Redis client", zap.Error(err))
	}
	defer redisClient.Close()

	// Создаем репозитории
	metricsRepo := adapters.NewRedisMetricsRepository(redisClient.GetClient(), zapLogger)
	dashboardRepo := adapters.NewRedisDashboardRepository(redisClient.GetClient(), zapLogger)
	metricsCollectorImpl := adapters.NewRedisMetricsCollector(redisClient.GetClient(), zapLogger)
	alertService := adapters.NewRedisAlertService(redisClient.GetClient(), zapLogger)

	// Создаем telemetry MetricsCollector
	telemetryMetricsCollector, err := telemetry.NewMetricsCollector(
		telemetryManager.GetMeter(),
		zapLogger,
	)
	if err != nil {
		zapLogger.Fatal("Failed to create telemetry metrics collector", zap.Error(err))
	}

	// Создаем сервисы
	analyticsService := services.NewAnalyticsService(
		metricsRepo,
		dashboardRepo,
		metricsCollectorImpl,
		alertService,
		zapLogger,
		telemetryMetricsCollector,
		tracingManager,
	)

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(createUnaryInterceptor(tracingManager, metricsCollector, zapLogger)),
		grpc.StreamInterceptor(createStreamInterceptor(tracingManager, metricsCollector, zapLogger)),
	)

	// Регистрируем сервисы
	// TODO: Здесь должна быть регистрация gRPC сервиса
	// pb.RegisterAnalyticsServiceServer(grpcServer, analyticsHandler)
	_ = analyticsService // Используется в будущих версиях

	// Регистрируем health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("analytics", grpc_health_v1.HealthCheckResponse_SERVING)

	// Включаем рефлексию для разработки
	reflection.Register(grpcServer)

	// Создаем HTTP сервер для метрик
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", healthCheckHandler(zapLogger))
	httpMux.HandleFunc("/ready", readinessHandler(redisClient, zapLogger))
	httpMux.HandleFunc("/metrics", metricsHandler(metricsCollector, zapLogger))

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      httpMux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем gRPC сервер
	grpcListener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		zapLogger.Fatal("Failed to listen on gRPC address",
			zap.String("address", cfg.GRPC.Address),
			zap.Error(err))
	}

	go func() {
		zapLogger.Info("Starting gRPC server", zap.String("address", cfg.GRPC.Address))
		if err := grpcServer.Serve(grpcListener); err != nil {
			zapLogger.Error("gRPC server failed", zap.Error(err))
			cancel()
		}
	}()

	// Запускаем HTTP сервер
	go func() {
		zapLogger.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Error("HTTP server failed", zap.Error(err))
			cancel()
		}
	}()

	// Запускаем метрики системы
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collectSystemMetrics(ctx, telemetryMetricsCollector, zapLogger)
			}
		}
	}()

	// Ждем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		zapLogger.Info("Application context cancelled")
	case sig := <-sigChan:
		zapLogger.Info("Received signal", zap.String("signal", sig.String()))
		cancel()
	}

	// Graceful shutdown
	zapLogger.Info("Shutting down servers...")

	// Устанавливаем статус NOT_SERVING
	healthServer.SetServingStatus("analytics", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Останавливаем gRPC сервер
	grpcServer.GracefulStop()

	// Останавливаем HTTP сервер
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	zapLogger.Info("Analytics service stopped")
}

// createUnaryInterceptor создает унарный интерсептор для gRPC
func createUnaryInterceptor(
	tracingManager *telemetry.TracingManager,
	metricsCollector *telemetry.MetricsCollector,
	logger *zap.Logger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Начинаем трейсинг
		ctx, span := tracingManager.StartTransaction(ctx, info.FullMethod)
		defer span.End()

		// Записываем метрики
		metricsCollector.RecordSystemRequest(ctx, "grpc", info.FullMethod)

		// Выполняем запрос
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		metricsCollector.RecordSystemResponseTime(ctx, duration, "grpc", info.FullMethod)

		if err != nil {
			metricsCollector.RecordSystemError(ctx, "grpc_error", "analytics")
			tracingManager.RecordError(ctx, err)
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Debug("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// createStreamInterceptor создает потоковый интерсептор для gRPC
func createStreamInterceptor(
	tracingManager *telemetry.TracingManager,
	metricsCollector *telemetry.MetricsCollector,
	logger *zap.Logger,
) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		ctx := ss.Context()

		// Начинаем трейсинг
		ctx, span := tracingManager.StartTransaction(ctx, info.FullMethod)
		defer span.End()

		// Записываем метрики
		metricsCollector.RecordSystemRequest(ctx, "grpc-stream", info.FullMethod)

		// Выполняем запрос
		err := handler(srv, ss)

		duration := time.Since(start)
		metricsCollector.RecordSystemResponseTime(ctx, duration, "grpc-stream", info.FullMethod)

		if err != nil {
			metricsCollector.RecordSystemError(ctx, "grpc_stream_error", "analytics")
			tracingManager.RecordError(ctx, err)
			logger.Error("gRPC stream request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Debug("gRPC stream request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return err
	}
}

// healthCheckHandler обработчик health check
func healthCheckHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"analytics","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}
}

// readinessHandler обработчик readiness check
func readinessHandler(redisClient *redis.Client, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Проверяем Redis
		if err := redisClient.Ping(ctx); err != nil {
			logger.Error("Redis health check failed", zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","error":"redis_unavailable","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"analytics","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}
}

// metricsHandler обработчик для экспорта метрик
func metricsHandler(metricsCollector *telemetry.MetricsCollector, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prometheus экспорт метрик происходит автоматически через OpenTelemetry
		// Этот handler может быть использован для дополнительной информации
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "# Metrics are exported via OpenTelemetry Prometheus exporter\n")
		fmt.Fprintf(w, "# See http://localhost:8889/metrics for Prometheus metrics\n")
	}
}

// collectSystemMetrics собирает системные метрики
func collectSystemMetrics(ctx context.Context, metricsCollector *telemetry.MetricsCollector, logger *zap.Logger) {
	// В реальном приложении здесь должен быть сбор реальных системных метрик

	// Имитируем системную нагрузку
	systemLoad := 0.5 + float64(time.Now().Unix()%10)/20.0
	metricsCollector.RecordSystemLoad(ctx, systemLoad)

	// Имитируем uptime
	uptime := time.Since(time.Now().Add(-time.Hour))
	metricsCollector.RecordSystemUptime(ctx, uptime)

	logger.Debug("System metrics collected",
		zap.Float64("system_load", systemLoad),
		zap.Duration("uptime", uptime),
	)
}
