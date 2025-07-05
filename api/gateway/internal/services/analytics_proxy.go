package services

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// AnalyticsProxy прокси для Analytics сервиса
type AnalyticsProxy struct {
	baseURL string
	logger  *zap.Logger
	client  *http.Client
}

// NewAnalyticsProxy создает новый прокси для Analytics
func NewAnalyticsProxy(baseURL string, logger *zap.Logger, client *http.Client) *AnalyticsProxy {
	return &AnalyticsProxy{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		logger:  logger,
		client:  client,
	}
}

// Proxy проксирует запрос к Analytics сервису
func (p *AnalyticsProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Убираем префикс /api/v1/analytics из пути
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/analytics")
	if path == "" {
		path = "/"
	}

	// Создаем новый URL для Analytics сервиса
	targetURL := p.baseURL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Создаем новый запрос
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		p.logger.Error("failed to create request", zap.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		p.logger.Error("failed to proxy request",
			zap.String("method", r.Method),
			zap.String("path", path),
			zap.String("error", err.Error()))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
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
		p.logger.Error("failed to copy response body", zap.String("error", err.Error()))
	}

	p.logger.Debug("proxied request to analytics",
		zap.String("method", r.Method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode))
}
