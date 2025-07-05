package services

import (
	"context"
	"fmt"
	"time"

	adaptershttp "github.com/par1ram/silence/rpc/analytics/internal/adapters/http"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
)

// MetricsCollectorImpl реализация сборщика метрик
type MetricsCollectorImpl struct {
	analyticsService ports.AnalyticsService
	logger           *zap.Logger
	stopChan         chan struct{}
	isRunning        bool
	client           adaptershttp.MetricsClient
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector(analyticsService ports.AnalyticsService, logger *zap.Logger, client adaptershttp.MetricsClient) ports.MetricsCollector {
	return &MetricsCollectorImpl{
		analyticsService: analyticsService,
		logger:           logger,
		stopChan:         make(chan struct{}),
		client:           client,
	}
}

// CollectConnectionMetrics собирает метрики подключений
func (c *MetricsCollectorImpl) CollectConnectionMetrics(ctx context.Context) error {
	metrics, err := c.client.GetConnectionMetrics(ctx)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := c.analyticsService.RecordConnection(ctx, metric); err != nil {
			c.logger.Warn("Failed to record connection metric", zap.String("error", err.Error()))
		}
	}
	return nil
}

// CollectBypassEffectivenessMetrics собирает метрики эффективности обхода DPI
func (c *MetricsCollectorImpl) CollectBypassEffectivenessMetrics(ctx context.Context) error {
	metrics, err := c.client.GetBypassEffectiveness(ctx)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := c.analyticsService.RecordBypassEffectiveness(ctx, metric); err != nil {
			c.logger.Warn("Failed to record bypass effectiveness metric", zap.String("error", err.Error()))
		}
	}
	return nil
}

// CollectUserActivityMetrics собирает метрики активности пользователей
func (c *MetricsCollectorImpl) CollectUserActivityMetrics(ctx context.Context) error {
	metrics, err := c.client.GetUserActivity(ctx)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := c.analyticsService.RecordUserActivity(ctx, metric); err != nil {
			c.logger.Warn("Failed to record user activity metric", zap.String("error", err.Error()))
		}
	}
	return nil
}

// CollectServerLoadMetrics собирает метрики нагрузки серверов
func (c *MetricsCollectorImpl) CollectServerLoadMetrics(ctx context.Context) error {
	metrics, err := c.client.GetServerLoad(ctx)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := c.analyticsService.RecordServerLoad(ctx, metric); err != nil {
			c.logger.Warn("Failed to record server load metric", zap.String("error", err.Error()))
		}
	}
	return nil
}

// CollectErrorMetrics собирает метрики ошибок
func (c *MetricsCollectorImpl) CollectErrorMetrics(ctx context.Context) error {
	// TODO: Реализовать сбор ошибок из сервисов (например, через отдельные эндпоинты или логи)
	c.logger.Debug("Collecting error metrics (not implemented)")
	return nil
}

// StartPeriodicCollection запускает периодический сбор метрик
func (c *MetricsCollectorImpl) StartPeriodicCollection(ctx context.Context) error {
	if c.isRunning {
		return fmt.Errorf("collection already running")
	}

	c.isRunning = true
	c.logger.Info("Starting periodic metrics collection")

	go func() {
		ticker := time.NewTicker(30 * time.Second) // Сбор каждые 30 секунд
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.collectAllMetrics(ctx)
			case <-c.stopChan:
				c.logger.Info("Stopping periodic metrics collection")
				return
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping collection")
				return
			}
		}
	}()

	return nil
}

// StopPeriodicCollection останавливает периодический сбор метрик
func (c *MetricsCollectorImpl) StopPeriodicCollection(ctx context.Context) error {
	if !c.isRunning {
		return fmt.Errorf("collection not running")
	}

	close(c.stopChan)
	c.isRunning = false
	return nil
}

// collectAllMetrics собирает все типы метрик
func (c *MetricsCollectorImpl) collectAllMetrics(ctx context.Context) {
	collectors := []func(context.Context) error{
		c.CollectConnectionMetrics,
		c.CollectBypassEffectivenessMetrics,
		c.CollectUserActivityMetrics,
		c.CollectServerLoadMetrics,
		c.CollectErrorMetrics,
	}

	for _, collector := range collectors {
		if err := collector(ctx); err != nil {
			c.logger.Error("Failed to collect metrics", zap.String("error", err.Error()))
		}
	}
}
