package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
)

// AnalyticsIntegration интеграция с сервисом аналитики
type AnalyticsIntegration struct {
	analyticsURL string
	httpClient   *http.Client
}

// NotificationDeliveryMetric метрика доставки уведомления
type NotificationDeliveryMetric struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	ErrorType   string            `json:"error_type"`
	Service     string            `json:"service"`
	UserID      string            `json:"user_id,omitempty"`
	ServerID    string            `json:"server_id,omitempty"`
	StatusCode  int               `json:"status_code,omitempty"`
	Description string            `json:"description"`
}

// NewAnalyticsIntegration создает новый интегратор с аналитикой
func NewAnalyticsIntegration(analyticsURL string) *AnalyticsIntegration {
	return &AnalyticsIntegration{
		analyticsURL: analyticsURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RecordNotificationDelivery записывает метрику успешной доставки
func (a *AnalyticsIntegration) RecordNotificationDelivery(ctx context.Context, notification *domain.Notification, channel domain.NotificationChannel) error {
	metric := NotificationDeliveryMetric{
		Name:      "notification_delivery_success",
		Type:      "counter",
		Value:     1.0,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"notification_type": string(notification.Type),
			"channel":           string(channel),
			"priority":          string(notification.Priority),
			"source":            notification.Source,
		},
		ErrorType:   "success",
		Service:     "notifications",
		UserID:      notification.Recipients[0], // берем первого получателя
		Description: fmt.Sprintf("Notification delivered via %s", channel),
	}

	return a.sendMetric(ctx, metric)
}

// RecordNotificationError записывает метрику ошибки доставки
func (a *AnalyticsIntegration) RecordNotificationError(ctx context.Context, notification *domain.Notification, channel domain.NotificationChannel, err error) error {
	metric := NotificationDeliveryMetric{
		Name:      "notification_delivery_error",
		Type:      "counter",
		Value:     1.0,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"notification_type": string(notification.Type),
			"channel":           string(channel),
			"priority":          string(notification.Priority),
			"source":            notification.Source,
		},
		ErrorType:   "delivery_failed",
		Service:     "notifications",
		UserID:      notification.Recipients[0], // берем первого получателя
		StatusCode:  500,
		Description: fmt.Sprintf("Failed to deliver notification via %s: %v", channel, err),
	}

	return a.sendMetric(ctx, metric)
}

// RecordNotificationStats записывает общую статистику уведомлений
func (a *AnalyticsIntegration) RecordNotificationStats(ctx context.Context, stats map[string]interface{}) error {
	metric := NotificationDeliveryMetric{
		Name:      "notification_stats",
		Type:      "gauge",
		Value:     float64(stats["total_sent"].(int64)),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"metric": "total_sent",
		},
		ErrorType:   "stats",
		Service:     "notifications",
		Description: "Notification delivery statistics",
	}

	return a.sendMetric(ctx, metric)
}

// sendMetric отправляет метрику в analytics сервис
func (a *AnalyticsIntegration) sendMetric(ctx context.Context, metric NotificationDeliveryMetric) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	url := fmt.Sprintf("%s/metrics/errors", a.analyticsURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Printf("[analytics] failed to send metric: %v", err)
		return fmt.Errorf("failed to send metric: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("[analytics] unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("[analytics] metric sent successfully: %s", metric.Name)
	return nil
}
