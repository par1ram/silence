package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"bytes"

	"go.uber.org/zap"
)

// DPIProxy проксирует запросы к DPI Bypass сервису
type DPIProxy struct {
	dpiBypassURL string
	logger       *zap.Logger
	client       *http.Client
}

// NewDPIProxy создает новый прокси для DPI Bypass сервиса
func NewDPIProxy(dpiBypassURL string, logger *zap.Logger, client *http.Client) *DPIProxy {
	return &DPIProxy{
		dpiBypassURL: dpiBypassURL,
		logger:       logger,
		client:       client,
	}
}

// Proxy проксирует запрос к DPI Bypass сервису
func (d *DPIProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для DPI Bypass сервиса
	dpiBypassURL, err := url.Parse(d.dpiBypassURL)
	if err != nil {
		d.logger.Error("failed to parse DPI Bypass URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(dpiBypassURL)

	// Настраиваем директор для изменения запроса
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Убираем префикс /api/v1/dpi-bypass из пути
		if req.URL.Path != "" {
			req.URL.Path = req.URL.Path[len("/api/v1/dpi-bypass"):]
		}

		d.logger.Info("proxying request to DPI Bypass service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		d.logger.Info("received response from DPI Bypass service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		d.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// CreateBypass создает bypass-конфигурацию через DPI Bypass сервис
func (d *DPIProxy) CreateBypass(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	dpiBypassURL, err := url.Parse(d.dpiBypassURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPI Bypass URL: %w", err)
	}

	bypassURL := dpiBypassURL.JoinPath("api", "v1", "bypass")

	// Сериализуем конфигурацию в JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bypass config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bypassURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create bypass request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create bypass: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create bypass: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode bypass response: %w", err)
	}

	return result, nil
}

// StartBypass запускает bypass-соединение
func (d *DPIProxy) StartBypass(ctx context.Context, id string) error {
	dpiBypassURL, err := url.Parse(d.dpiBypassURL)
	if err != nil {
		return fmt.Errorf("failed to parse DPI Bypass URL: %w", err)
	}

	startURL := dpiBypassURL.JoinPath("api", "v1", "bypass", id, "start")

	req, err := http.NewRequestWithContext(ctx, "POST", startURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create start bypass request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start bypass: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start bypass: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}
