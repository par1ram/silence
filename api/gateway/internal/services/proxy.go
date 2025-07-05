package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"bytes"

	"go.uber.org/zap"
)

// ProxyService сервис для проксирования запросов
type ProxyService struct {
	authURL      string
	vpnCoreURL   string
	dpiBypassURL string
	logger       *zap.Logger
	client       *http.Client
}

// NewProxyService создает новый сервис проксирования
func NewProxyService(authURL, vpnCoreURL, dpiBypassURL string, logger *zap.Logger) *ProxyService {
	return &ProxyService{
		authURL:      authURL,
		vpnCoreURL:   vpnCoreURL,
		dpiBypassURL: dpiBypassURL,
		logger:       logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToAuth проксирует запрос к auth сервису
func (p *ProxyService) ProxyToAuth(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для auth сервиса
	authURL, err := url.Parse(p.authURL)
	if err != nil {
		p.logger.Error("failed to parse auth URL", zap.Error(err))
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

		p.logger.Info("proxying request to auth service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		p.logger.Info("received response from auth service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// ProxyToVPNCore проксирует запрос к VPN Core сервису
func (p *ProxyService) ProxyToVPNCore(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для VPN Core сервиса
	vpnCoreURL, err := url.Parse(p.vpnCoreURL)
	if err != nil {
		p.logger.Error("failed to parse VPN Core URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(vpnCoreURL)

	// Настраиваем директор для изменения запроса
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Убираем префикс /api/v1/vpn из пути
		if req.URL.Path != "" {
			req.URL.Path = req.URL.Path[len("/api/v1/vpn"):]
		}

		p.logger.Info("proxying request to VPN Core service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		p.logger.Info("received response from VPN Core service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// HealthCheck проверяет доступность auth сервиса
func (p *ProxyService) HealthCheck(ctx context.Context) error {
	authURL, err := url.Parse(p.authURL)
	if err != nil {
		return fmt.Errorf("failed to parse auth URL: %w", err)
	}

	healthURL := authURL.JoinPath("health")

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
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

// ProxyToDPIBypass проксирует запрос к DPI Bypass сервису
func (p *ProxyService) ProxyToDPIBypass(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для DPI Bypass сервиса
	dpiBypassURL, err := url.Parse(p.dpiBypassURL)
	if err != nil {
		p.logger.Error("failed to parse DPI Bypass URL", zap.Error(err))
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

		p.logger.Info("proxying request to DPI Bypass service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		p.logger.Info("received response from DPI Bypass service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// CreateBypass создает bypass-конфигурацию через DPI Bypass сервис
func (p *ProxyService) CreateBypass(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	dpiBypassURL, err := url.Parse(p.dpiBypassURL)
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

	resp, err := p.client.Do(req)
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
func (p *ProxyService) StartBypass(ctx context.Context, id string) error {
	dpiBypassURL, err := url.Parse(p.dpiBypassURL)
	if err != nil {
		return fmt.Errorf("failed to parse DPI Bypass URL: %w", err)
	}

	startURL := dpiBypassURL.JoinPath("api", "v1", "bypass", id, "start")

	req, err := http.NewRequestWithContext(ctx, "POST", startURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create start bypass request: %w", err)
	}

	resp, err := p.client.Do(req)
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

// CreateVPNTunnel создает VPN-туннель через VPN Core сервис
func (p *ProxyService) CreateVPNTunnel(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	vpnCoreURL, err := url.Parse(p.vpnCoreURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VPN Core URL: %w", err)
	}

	tunnelURL := vpnCoreURL.JoinPath("tunnels")

	// Сериализуем конфигурацию в JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VPN config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tunnelURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create VPN tunnel request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create VPN tunnel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create VPN tunnel: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode VPN tunnel response: %w", err)
	}

	return result, nil
}
