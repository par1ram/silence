package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/ports"
	"go.uber.org/zap"
)

// AlertServiceImpl реализация сервиса уведомлений
type AlertServiceImpl struct {
	analyticsService ports.AnalyticsService
	logger           *zap.Logger
	alerts           map[string]*domain.AlertRule
}

// NewAlertService создает новый сервис уведомлений
func NewAlertService(analyticsService ports.AnalyticsService, logger *zap.Logger) ports.AlertService {
	return &AlertServiceImpl{
		analyticsService: analyticsService,
		logger:           logger,
		alerts:           make(map[string]*domain.AlertRule),
	}
}

// CreateAlertRule создает правило уведомления
func (a *AlertServiceImpl) CreateAlertRule(ctx context.Context, rule domain.AlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("alert rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("alert rule name is required")
	}

	if rule.Condition == "" {
		return fmt.Errorf("alert rule condition is required")
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.Status = domain.AlertStatusActive

	a.alerts[rule.ID] = &rule
	a.logger.Info("Created alert rule", zap.String("id", rule.ID), zap.String("name", rule.Name))

	return nil
}

// GetAlertRule получает правило уведомления
func (a *AlertServiceImpl) GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error) {
	rule, exists := a.alerts[id]
	if !exists {
		return nil, fmt.Errorf("alert rule not found: %s", id)
	}

	return rule, nil
}

// UpdateAlertRule обновляет правило уведомления
func (a *AlertServiceImpl) UpdateAlertRule(ctx context.Context, rule domain.AlertRule) error {
	if _, exists := a.alerts[rule.ID]; !exists {
		return fmt.Errorf("alert rule not found: %s", rule.ID)
	}

	rule.UpdatedAt = time.Now()
	a.alerts[rule.ID] = &rule

	a.logger.Info("Updated alert rule", zap.String("id", rule.ID))
	return nil
}

// DeleteAlertRule удаляет правило уведомления
func (a *AlertServiceImpl) DeleteAlertRule(ctx context.Context, id string) error {
	if _, exists := a.alerts[id]; !exists {
		return fmt.Errorf("alert rule not found: %s", id)
	}

	delete(a.alerts, id)
	a.logger.Info("Deleted alert rule", zap.String("id", id))

	return nil
}

// ListAlertRules получает список правил уведомлений
func (a *AlertServiceImpl) ListAlertRules(ctx context.Context) ([]domain.AlertRule, error) {
	var rules []domain.AlertRule
	for _, rule := range a.alerts {
		rules = append(rules, *rule)
	}

	return rules, nil
}

// EvaluateAlerts оценивает все активные правила уведомлений
func (a *AlertServiceImpl) EvaluateAlerts(ctx context.Context) error {
	for _, rule := range a.alerts {
		if rule.Status != domain.AlertStatusActive {
			continue
		}

		if err := a.evaluateRule(ctx, rule); err != nil {
			a.logger.Error("Failed to evaluate alert rule",
				zap.String("id", rule.ID),
				zap.String("error", err.Error()))
		}
	}

	return nil
}

// evaluateRule оценивает конкретное правило
func (a *AlertServiceImpl) evaluateRule(ctx context.Context, rule *domain.AlertRule) error {
	// TODO: Реализовать логику оценки условий
	// Пока что просто логируем
	a.logger.Debug("Evaluating alert rule",
		zap.String("id", rule.ID),
		zap.String("condition", rule.Condition))

	// Пример оценки: если условие содержит "high_load"
	if rule.Condition == "high_load" {
		// Проверяем метрики нагрузки серверов
		metrics, err := a.analyticsService.GetServerLoadMetrics(ctx, time.Now().Add(-5*time.Minute), time.Now())
		if err != nil {
			return fmt.Errorf("failed to get server load metrics: %w", err)
		}

		for _, metric := range metrics {
			if metric.CPUUsage > 90.0 || metric.MemoryUsage > 90.0 {
				return a.triggerAlert(ctx, rule, metric)
			}
		}
	}

	return nil
}

// triggerAlert срабатывает уведомление
func (a *AlertServiceImpl) triggerAlert(ctx context.Context, rule *domain.AlertRule, metric domain.ServerLoadMetric) error {
	_ = domain.Alert{
		ID:          fmt.Sprintf("alert_%d", time.Now().Unix()),
		RuleID:      rule.ID,
		Severity:    rule.Severity,
		Message:     rule.Message,
		Status:      domain.AlertStatusTriggered,
		CreatedAt:   time.Now(),
		MetricValue: metric.CPUUsage,
		ServerID:    metric.ServerID,
	}

	// TODO: Отправить уведомление через notification service
	a.logger.Warn("Alert triggered",
		zap.String("rule_id", rule.ID),
		zap.String("severity", string(rule.Severity)),
		zap.String("message", rule.Message))

	return nil
}

// GetAlertHistory получает историю уведомлений
func (a *AlertServiceImpl) GetAlertHistory(ctx context.Context, ruleID string, limit int) ([]domain.Alert, error) {
	// TODO: Реализовать получение из базы данных
	// Пока возвращаем пустой список
	return []domain.Alert{}, nil
}

// AcknowledgeAlert подтверждает уведомление
func (a *AlertServiceImpl) AcknowledgeAlert(ctx context.Context, alertID string) error {
	// TODO: Реализовать подтверждение уведомления
	a.logger.Info("Alert acknowledged", zap.String("alert_id", alertID))
	return nil
}

// ResolveAlert разрешает уведомление
func (a *AlertServiceImpl) ResolveAlert(ctx context.Context, alertID string) error {
	// TODO: Реализовать разрешение уведомления
	a.logger.Info("Alert resolved", zap.String("alert_id", alertID))
	return nil
}
