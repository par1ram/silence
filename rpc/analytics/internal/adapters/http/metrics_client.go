package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
)

// MetricsClient интерфейс для получения метрик из других сервисов
type MetricsClient interface {
	GetServerLoad(ctx context.Context) ([]domain.ServerLoadMetric, error)
	GetConnectionMetrics(ctx context.Context) ([]domain.ConnectionMetric, error)
	GetBypassEffectiveness(ctx context.Context) ([]domain.BypassEffectivenessMetric, error)
	GetUserActivity(ctx context.Context) ([]domain.UserActivityMetric, error)
}

// HTTPMetricsClient реализация MetricsClient через HTTP
// URL каждого сервиса передается при создании

type HTTPMetricsClient struct {
	VPNCoreURL string
	GatewayURL string
	BypassURL  string
	AuthURL    string
	client     *http.Client
}

func NewHTTPMetricsClient(vpnCore, gateway, bypass, auth string) *HTTPMetricsClient {
	return &HTTPMetricsClient{
		VPNCoreURL: vpnCore,
		GatewayURL: gateway,
		BypassURL:  bypass,
		AuthURL:    auth,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *HTTPMetricsClient) GetServerLoad(ctx context.Context) ([]domain.ServerLoadMetric, error) {
	return getMetrics[domain.ServerLoadMetric](ctx, c.client, c.VPNCoreURL+"/metrics/server-load")
}

func (c *HTTPMetricsClient) GetConnectionMetrics(ctx context.Context) ([]domain.ConnectionMetric, error) {
	return getMetrics[domain.ConnectionMetric](ctx, c.client, c.GatewayURL+"/metrics/connections")
}

func (c *HTTPMetricsClient) GetBypassEffectiveness(ctx context.Context) ([]domain.BypassEffectivenessMetric, error) {
	return getMetrics[domain.BypassEffectivenessMetric](ctx, c.client, c.BypassURL+"/metrics/bypass-effectiveness")
}

func (c *HTTPMetricsClient) GetUserActivity(ctx context.Context) ([]domain.UserActivityMetric, error) {
	return getMetrics[domain.UserActivityMetric](ctx, c.client, c.AuthURL+"/metrics/user-activity")
}

// Универсальный метод для получения метрик по HTTP
func getMetrics[T any](ctx context.Context, client *http.Client, url string) ([]T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var metrics []T
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}
	return metrics, nil
}
