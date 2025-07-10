package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NotificationsProxy проксирует запросы к Notifications сервису
type NotificationsProxy struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewNotificationsProxy создает новый прокси для Notifications сервиса
func NewNotificationsProxy(baseURL string, logger *zap.Logger, client *http.Client) *NotificationsProxy {
	return &NotificationsProxy{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

// Proxy проксирует HTTP запрос к Notifications сервису
func (p *NotificationsProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Логируем входящий запрос
	p.logger.Info("proxying request to notifications service",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Подготавливаем URL для проксирования
	targetURL := p.baseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Читаем тело запроса, если оно есть
	var body io.Reader
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			p.logger.Error("failed to read request body", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Создаем новый запрос
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, body)
	if err != nil {
		p.logger.Error("failed to create request", zap.Error(err))
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Копируем заголовки
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Добавляем заголовки для внутреннего взаимодействия
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("X-Forwarded-Proto", "http")
	if r.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	}

	// Выполняем запрос
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("failed to proxy request", zap.Error(err))
		http.Error(w, "Notifications service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Устанавливаем код статуса
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	if _, err := io.Copy(w, resp.Body); err != nil {
		p.logger.Error("failed to copy response body", zap.Error(err))
		return
	}

	p.logger.Info("successfully proxied request to notifications service",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", resp.StatusCode),
	)
}

// HealthCheck проверяет доступность Notifications сервиса
func (p *NotificationsProxy) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("notifications service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notifications service health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// DispatchNotification отправляет уведомление через Notifications сервис
func (p *NotificationsProxy) DispatchNotification(ctx context.Context, notification map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/v1/notifications/dispatch", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create dispatch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("notification dispatch failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetNotification получает уведомление по ID
func (p *NotificationsProxy) GetNotification(ctx context.Context, id string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/v1/notifications/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get notification failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// ListNotifications получает список уведомлений
func (p *NotificationsProxy) ListNotifications(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/v1/notifications", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	// Добавляем query parameters из фильтров
	if len(filters) > 0 {
		q := req.URL.Query()
		for key, value := range filters {
			q.Add(key, fmt.Sprintf("%v", value))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list notifications failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CreateTemplate создает шаблон уведомления
func (p *NotificationsProxy) CreateTemplate(ctx context.Context, template map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/v1/notifications/templates", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create template request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("template creation failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// UpdateUserPreferences обновляет настройки уведомлений пользователя
func (p *NotificationsProxy) UpdateUserPreferences(ctx context.Context, userID string, preferences map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", p.baseURL+"/api/v1/notifications/preferences/"+userID, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create preferences request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("preferences update failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetNotificationStats получает статистику уведомлений
func (p *NotificationsProxy) GetNotificationStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/v1/notifications/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats request: %w", err)
	}

	// Добавляем query parameters из фильтров
	if len(filters) > 0 {
		q := req.URL.Query()
		for key, value := range filters {
			q.Add(key, fmt.Sprintf("%v", value))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get stats failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
