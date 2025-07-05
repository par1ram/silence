package services

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// ServerManagerProxy прокси для Server Manager сервиса
type ServerManagerProxy struct {
	baseURL string
	logger  *zap.Logger
	client  *http.Client
}

// NewServerManagerProxy создает новый прокси для Server Manager
func NewServerManagerProxy(baseURL string, logger *zap.Logger, client *http.Client) *ServerManagerProxy {
	return &ServerManagerProxy{
		baseURL: baseURL,
		logger:  logger,
		client:  client,
	}
}

// Proxy проксирует запрос к Server Manager сервису
func (p *ServerManagerProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Создаем новый URL для запроса
	targetURL := p.baseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Создаем новый запрос
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		p.logger.Error("failed to create request", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Копируем заголовки
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Выполняем запрос
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("failed to proxy request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Устанавливаем статус код
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	if _, err := io.Copy(w, resp.Body); err != nil {
		p.logger.Error("failed to copy response body", zap.Error(err))
	}
}

// HealthCheck проверяет доступность Server Manager сервиса
func (p *ServerManagerProxy) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
	}

	return nil
}
