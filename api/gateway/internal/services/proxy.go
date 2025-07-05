package services

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ProxyService сервис для проксирования запросов
type ProxyService struct {
	authProxy          *AuthProxy
	vpnProxy           *VPNProxy
	dpiProxy           *DPIProxy
	analyticsProxy     *AnalyticsProxy
	serverManagerProxy *ServerManagerProxy
	logger             *zap.Logger
}

// NewProxyService создает новый сервис проксирования
func NewProxyService(authURL, vpnCoreURL, dpiBypassURL, analyticsURL, serverManagerURL, internalToken string, logger *zap.Logger) *ProxyService {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &ProxyService{
		authProxy:          NewAuthProxy(authURL, internalToken, logger, client),
		vpnProxy:           NewVPNProxy(vpnCoreURL, logger, client),
		dpiProxy:           NewDPIProxy(dpiBypassURL, logger, client),
		analyticsProxy:     NewAnalyticsProxy(analyticsURL, logger, client),
		serverManagerProxy: NewServerManagerProxy(serverManagerURL, logger, client),
		logger:             logger,
	}
}

// ProxyToAuth проксирует запрос к auth сервису
func (p *ProxyService) ProxyToAuth(w http.ResponseWriter, r *http.Request) {
	p.authProxy.Proxy(w, r)
}

// ProxyToVPNCore проксирует запрос к VPN Core сервису
func (p *ProxyService) ProxyToVPNCore(w http.ResponseWriter, r *http.Request) {
	p.vpnProxy.Proxy(w, r)
}

// ProxyToDPIBypass проксирует запрос к DPI Bypass сервису
func (p *ProxyService) ProxyToDPIBypass(w http.ResponseWriter, r *http.Request) {
	p.dpiProxy.Proxy(w, r)
}

// ProxyToAnalytics проксирует запрос к Analytics сервису
func (p *ProxyService) ProxyToAnalytics(w http.ResponseWriter, r *http.Request) {
	p.analyticsProxy.Proxy(w, r)
}

// ProxyToServerManager проксирует запрос к Server Manager сервису
func (p *ProxyService) ProxyToServerManager(w http.ResponseWriter, r *http.Request) {
	p.serverManagerProxy.Proxy(w, r)
}

// HealthCheck проверяет доступность auth сервиса
func (p *ProxyService) HealthCheck(ctx context.Context) error {
	return p.authProxy.HealthCheck(ctx)
}

// CreateBypass создает bypass-конфигурацию через DPI Bypass сервис
func (p *ProxyService) CreateBypass(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	return p.dpiProxy.CreateBypass(ctx, config)
}

// StartBypass запускает bypass-соединение
func (p *ProxyService) StartBypass(ctx context.Context, id string) error {
	return p.dpiProxy.StartBypass(ctx, id)
}

// CreateVPNTunnel создает VPN-туннель через VPN Core сервис
func (p *ProxyService) CreateVPNTunnel(ctx context.Context, config map[string]interface{}) (map[string]interface{}, error) {
	return p.vpnProxy.CreateTunnel(ctx, config)
}
