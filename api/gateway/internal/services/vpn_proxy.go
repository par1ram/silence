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

// VPNProxy проксирует запросы к VPN Core сервису
type VPNProxy struct {
	vpnCoreURL string
	logger     *zap.Logger
	client     *http.Client
}

// NewVPNProxy создает новый прокси для VPN Core сервиса
func NewVPNProxy(vpnCoreURL string, logger *zap.Logger, client *http.Client) *VPNProxy {
	return &VPNProxy{
		vpnCoreURL: vpnCoreURL,
		logger:     logger,
		client:     client,
	}
}

// Proxy проксирует запрос к VPN Core сервису
func (v *VPNProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	// Создаем URL для VPN Core сервиса
	vpnCoreURL, err := url.Parse(v.vpnCoreURL)
	if err != nil {
		v.logger.Error("failed to parse VPN Core URL", zap.Error(err))
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

		v.logger.Info("proxying request to VPN Core service",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("target", req.URL.String()))
	}

	// Настраиваем модификатор ответа
	proxy.ModifyResponse = func(resp *http.Response) error {
		v.logger.Info("received response from VPN Core service",
			zap.Int("status", resp.StatusCode),
			zap.String("path", r.URL.Path))
		return nil
	}

	// Обрабатываем ошибки
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		v.logger.Error("proxy error", zap.Error(err))
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

// CreateTunnel создает VPN-туннель через VPN Core сервис
func (v *VPNProxy) CreateTunnel(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	vpnCoreURL, err := url.Parse(v.vpnCoreURL)
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

	resp, err := v.client.Do(req)
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
