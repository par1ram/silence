package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

// AuthProxy проксирует запросы к auth сервису
type AuthProxy struct {
	authURL       string
	internalToken string
	logger        *zap.Logger
	client        *http.Client
}

// NewAuthProxy создает новый прокси для auth сервиса
func NewAuthProxy(authURL, internalToken string, logger *zap.Logger, client *http.Client) *AuthProxy {
	return &AuthProxy{
		authURL:       authURL,
		internalToken: internalToken,
		logger:        logger,
		client:        client,
	}
}

// Proxy проксирует запрос к auth сервису
func (a *AuthProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для auth сервиса
	authURL, err := url.Parse(a.authURL)
	if err != nil {
		a.logger.Error("failed to parse auth URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(authURL)

	// Настраиваем директор для изменения запроса
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Убираем префикс /api/v1/auth из пути
		if req.URL.Path != "" {
			req.URL.Path = req.URL.Path[len("/api/v1/auth"):]
		}

		// Добавляем внутренний токен для доступа к auth-сервису
		req.Header.Set("X-Internal-Token", a.internalToken)

		a.logger.Info("proxying request to auth service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		a.logger.Info("received response from auth service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		a.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// HealthCheck проверяет доступность auth сервиса
func (a *AuthProxy) HealthCheck(ctx context.Context) error {
	authURL, err := url.Parse(a.authURL)
	if err != nil {
		return fmt.Errorf("failed to parse auth URL: %w", err)
	}

	healthURL := authURL.JoinPath("health")

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Добавляем внутренний токен для health check
	req.Header.Set("X-Internal-Token", a.internalToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check auth service health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth service health check failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}
