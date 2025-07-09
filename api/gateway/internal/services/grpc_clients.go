package services

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	authClient "github.com/par1ram/silence/api/gateway/internal/clients/auth"
	notificationsClient "github.com/par1ram/silence/api/gateway/internal/clients/notifications"
	"github.com/par1ram/silence/api/gateway/internal/config"
)

// GRPCClients менеджер для gRPC клиентов
type GRPCClients struct {
	Auth          *authClient.Client
	Notifications *notificationsClient.Client
	// Analytics     *analyticsClient.Client
	// DPIBypass     *dpiBypassClient.Client
	// ServerManager *serverManagerClient.Client
	// VPNCore       *vpnCoreClient.Client

	mu     sync.RWMutex
	config *config.Config
	logger *zap.Logger
}

// NewGRPCClients создает новый менеджер gRPC клиентов
func NewGRPCClients(cfg *config.Config, logger *zap.Logger) *GRPCClients {
	return &GRPCClients{
		config: cfg,
		logger: logger,
	}
}

// Initialize инициализирует все gRPC клиенты
func (g *GRPCClients) Initialize(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Инициализируем auth клиент
	if err := g.initAuthClient(); err != nil {
		return fmt.Errorf("failed to initialize auth client: %w", err)
	}

	// Инициализируем notifications клиент
	if err := g.initNotificationsClient(); err != nil {
		return fmt.Errorf("failed to initialize notifications client: %w", err)
	}

	// TODO: Инициализировать остальные клиенты
	// Analytics, DPIBypass, ServerManager, VPNCore

	g.logger.Info("All gRPC clients initialized successfully")
	return nil
}

// initAuthClient инициализирует auth клиент
func (g *GRPCClients) initAuthClient() error {
	if g.config.AuthGRPCURL == "" {
		return fmt.Errorf("auth gRPC URL is not configured")
	}

	client, err := authClient.NewClient(g.config.AuthGRPCURL)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	g.Auth = client
	g.logger.Info("Auth gRPC client initialized", zap.String("url", g.config.AuthGRPCURL))
	return nil
}

// initNotificationsClient инициализирует notifications клиент
func (g *GRPCClients) initNotificationsClient() error {
	if g.config.NotificationsGRPCURL == "" {
		return fmt.Errorf("notifications gRPC URL is not configured")
	}

	client, err := notificationsClient.NewClient(g.config.NotificationsGRPCURL)
	if err != nil {
		return fmt.Errorf("failed to create notifications client: %w", err)
	}

	g.Notifications = client
	g.logger.Info("Notifications gRPC client initialized", zap.String("url", g.config.NotificationsGRPCURL))
	return nil
}

// Close закрывает все gRPC соединения
func (g *GRPCClients) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var errors []error

	if g.Auth != nil {
		if err := g.Auth.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close auth client: %w", err))
		}
	}

	if g.Notifications != nil {
		if err := g.Notifications.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close notifications client: %w", err))
		}
	}

	// TODO: Закрыть остальные клиенты

	if len(errors) > 0 {
		return fmt.Errorf("errors closing gRPC clients: %v", errors)
	}

	g.logger.Info("All gRPC clients closed successfully")
	return nil
}

// HealthCheck проверяет здоровье всех gRPC сервисов
func (g *GRPCClients) HealthCheck(ctx context.Context) map[string]error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	results := make(map[string]error)

	// Проверяем auth сервис
	if g.Auth != nil {
		_, err := g.Auth.Health(ctx)
		results["auth"] = err
	} else {
		results["auth"] = fmt.Errorf("auth client not initialized")
	}

	// Проверяем notifications сервис
	if g.Notifications != nil {
		_, err := g.Notifications.Health(ctx)
		results["notifications"] = err
	} else {
		results["notifications"] = fmt.Errorf("notifications client not initialized")
	}

	// TODO: Добавить проверки для остальных сервисов

	return results
}

// GetAuth возвращает auth клиент (потокобезопасно)
func (g *GRPCClients) GetAuth() *authClient.Client {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Auth
}

// GetNotifications возвращает notifications клиент (потокобезопасно)
func (g *GRPCClients) GetNotifications() *notificationsClient.Client {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Notifications
}

// IsReady проверяет, готовы ли все клиенты к работе
func (g *GRPCClients) IsReady() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.Auth != nil && g.Notifications != nil
	// TODO: Добавить проверки для остальных клиентов
}
