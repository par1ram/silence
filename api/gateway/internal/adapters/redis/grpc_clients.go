package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	authClientPkg "github.com/par1ram/silence/api/gateway/internal/clients/auth"
	notificationsClientPkg "github.com/par1ram/silence/api/gateway/internal/clients/notifications"
	"github.com/par1ram/silence/api/gateway/internal/config"
	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// RedisGRPCClients Redis-based менеджер для gRPC клиентов
type RedisGRPCClients struct {
	redisClient *sharedRedis.Client
	logger      *zap.Logger
	config      *RedisGRPCClientsConfig

	// Локальный кэш для активных соединений
	clients map[string]interface{}
	mu      sync.RWMutex
}

// RedisGRPCClientsConfig конфигурация Redis GRPC клиентов
type RedisGRPCClientsConfig struct {
	KeyPrefix      string
	HealthCheckTTL time.Duration
	ConnectionTTL  time.Duration
	RetryInterval  time.Duration
	MaxRetries     int
	CircuitBreaker bool
	LoadBalancing  bool
	EndpointURLs   map[string][]string
}

// ClientHealthInfo информация о здоровье клиента
type ClientHealthInfo struct {
	ServiceName  string    `json:"service_name"`
	Endpoint     string    `json:"endpoint"`
	IsHealthy    bool      `json:"is_healthy"`
	LastCheck    time.Time `json:"last_check"`
	LastError    string    `json:"last_error,omitempty"`
	ResponseTime int64     `json:"response_time_ms"`
	FailureCount int       `json:"failure_count"`
	CircuitOpen  bool      `json:"circuit_open"`
}

// ServiceEndpoint информация о endpoint сервиса
type ServiceEndpoint struct {
	URL            string    `json:"url"`
	Priority       int       `json:"priority"`
	Weight         int       `json:"weight"`
	IsActive       bool      `json:"is_active"`
	LastUsed       time.Time `json:"last_used"`
	FailureCount   int       `json:"failure_count"`
	AverageLatency int64     `json:"average_latency_ms"`
}

// ClientStats статистика использования клиента
type ClientStats struct {
	ServiceName        string    `json:"service_name"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	AverageLatency     int64     `json:"average_latency_ms"`
	LastRequestTime    time.Time `json:"last_request_time"`
	ConnectionsCount   int       `json:"connections_count"`
}

// NewRedisGRPCClients создает новый Redis-based менеджер gRPC клиентов
func NewRedisGRPCClients(redisClient *sharedRedis.Client, config *RedisGRPCClientsConfig, logger *zap.Logger) *RedisGRPCClients {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "gateway:grpc_clients"
	}
	if config.HealthCheckTTL == 0 {
		config.HealthCheckTTL = 30 * time.Second
	}
	if config.ConnectionTTL == 0 {
		config.ConnectionTTL = 5 * time.Minute
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.EndpointURLs == nil {
		config.EndpointURLs = make(map[string][]string)
	}

	manager := &RedisGRPCClients{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
		clients:     make(map[string]interface{}),
	}

	// Запускаем фоновые процессы
	go manager.startHealthCheckRoutine()
	go manager.startStatsCollectionRoutine()

	return manager
}

// InitializeClients инициализирует все gRPC клиенты
func (r *RedisGRPCClients) InitializeClients(ctx context.Context, appConfig *config.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Инициализируем auth клиент
	if err := r.initAuthClient(ctx, appConfig.AuthGRPCURL); err != nil {
		return fmt.Errorf("failed to initialize auth client: %w", err)
	}

	// Инициализируем notifications клиент
	if err := r.initNotificationsClient(ctx, appConfig.NotificationsGRPCURL); err != nil {
		return fmt.Errorf("failed to initialize notifications client: %w", err)
	}

	r.logger.Info("All gRPC clients initialized successfully")
	return nil
}

// initAuthClient инициализирует auth клиент
func (r *RedisGRPCClients) initAuthClient(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("auth gRPC URL is not configured")
	}

	// Проверяем доступность endpoint
	if !r.isEndpointHealthy(ctx, "auth", url) {
		// Пытаемся найти альтернативный endpoint
		if alternativeURL := r.getHealthyEndpoint(ctx, "auth"); alternativeURL != "" {
			url = alternativeURL
		}
	}

	client, err := authClientPkg.NewClient(url)
	if err != nil {
		r.recordFailure(ctx, "auth", url, err)
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	r.clients["auth"] = client
	r.recordSuccess(ctx, "auth", url)
	r.logger.Info("Auth gRPC client initialized", zap.String("url", url))
	return nil
}

// initNotificationsClient инициализирует notifications клиент
func (r *RedisGRPCClients) initNotificationsClient(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("notifications gRPC URL is not configured")
	}

	// Проверяем доступность endpoint
	if !r.isEndpointHealthy(ctx, "notifications", url) {
		// Пытаемся найти альтернативный endpoint
		if alternativeURL := r.getHealthyEndpoint(ctx, "notifications"); alternativeURL != "" {
			url = alternativeURL
		}
	}

	client, err := notificationsClientPkg.NewClient(url)
	if err != nil {
		r.recordFailure(ctx, "notifications", url, err)
		return fmt.Errorf("failed to create notifications client: %w", err)
	}

	r.clients["notifications"] = client
	r.recordSuccess(ctx, "notifications", url)
	r.logger.Info("Notifications gRPC client initialized", zap.String("url", url))
	return nil
}

// GetAuthClient возвращает auth клиент с автоматическим переподключением
func (r *RedisGRPCClients) GetAuthClient(ctx context.Context) (*authClientPkg.Client, error) {
	r.mu.RLock()
	client, exists := r.clients["auth"]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("auth client not initialized")
	}

	authClient, ok := client.(*authClientPkg.Client)
	if !ok {
		return nil, fmt.Errorf("invalid auth client type")
	}

	// Проверяем здоровье соединения
	if !r.isClientHealthy(ctx, "auth", authClient) {
		// Пытаемся переподключиться
		if err := r.reconnectClient(ctx, "auth"); err != nil {
			return nil, fmt.Errorf("failed to reconnect auth client: %w", err)
		}

		// Получаем новый клиент после переподключения
		r.mu.RLock()
		client, exists = r.clients["auth"]
		r.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("auth client not available after reconnection")
		}

		authClient, ok = client.(*authClientPkg.Client)
		if !ok {
			return nil, fmt.Errorf("invalid auth client type after reconnection")
		}
	}

	r.updateClientStats(ctx, "auth")
	return authClient, nil
}

// GetNotificationsClient возвращает notifications клиент с автоматическим переподключением
func (r *RedisGRPCClients) GetNotificationsClient(ctx context.Context) (*notificationsClientPkg.Client, error) {
	r.mu.RLock()
	client, exists := r.clients["notifications"]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("notifications client not initialized")
	}

	notifClient, ok := client.(*notificationsClientPkg.Client)
	if !ok {
		return nil, fmt.Errorf("invalid notifications client type")
	}

	// Проверяем здоровье соединения
	if !r.isClientHealthy(ctx, "notifications", notifClient) {
		// Пытаемся переподключиться
		if err := r.reconnectClient(ctx, "notifications"); err != nil {
			return nil, fmt.Errorf("failed to reconnect notifications client: %w", err)
		}

		// Получаем новый клиент после переподключения
		r.mu.RLock()
		client, exists = r.clients["notifications"]
		r.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("notifications client not available after reconnection")
		}

		notifClient, ok = client.(*notificationsClientPkg.Client)
		if !ok {
			return nil, fmt.Errorf("invalid notifications client type after reconnection")
		}
	}

	r.updateClientStats(ctx, "notifications")
	return notifClient, nil
}

// isEndpointHealthy проверяет здоровье endpoint
func (r *RedisGRPCClients) isEndpointHealthy(ctx context.Context, serviceName, endpoint string) bool {
	healthKey := fmt.Sprintf("%s:health:%s:%s", r.config.KeyPrefix, serviceName, endpoint)

	var healthInfo ClientHealthInfo
	if err := r.redisClient.Get(ctx, healthKey, &healthInfo); err != nil {
		return false
	}

	// Проверяем актуальность информации
	if time.Since(healthInfo.LastCheck) > r.config.HealthCheckTTL {
		return false
	}

	return healthInfo.IsHealthy && !healthInfo.CircuitOpen
}

// getHealthyEndpoint возвращает здоровый endpoint для сервиса
func (r *RedisGRPCClients) getHealthyEndpoint(ctx context.Context, serviceName string) string {
	endpoints := r.config.EndpointURLs[serviceName]
	if len(endpoints) == 0 {
		return ""
	}

	// Сортируем endpoints по приоритету и здоровью
	for _, endpoint := range endpoints {
		if r.isEndpointHealthy(ctx, serviceName, endpoint) {
			return endpoint
		}
	}

	return ""
}

// isClientHealthy проверяет здоровье клиента
func (r *RedisGRPCClients) isClientHealthy(ctx context.Context, serviceName string, client interface{}) bool {
	switch serviceName {
	case "auth":
		authClient, ok := client.(*authClientPkg.Client)
		if !ok {
			return false
		}
		_, err := authClient.Health(ctx)
		return err == nil
	case "notifications":
		notifClient, ok := client.(*notificationsClientPkg.Client)
		if !ok {
			return false
		}
		_, err := notifClient.Health(ctx)
		return err == nil
	default:
		return false
	}
}

// reconnectClient переподключает клиент
func (r *RedisGRPCClients) reconnectClient(ctx context.Context, serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Закрываем старое соединение
	if client, exists := r.clients[serviceName]; exists {
		switch serviceName {
		case "auth":
			if authClient, ok := client.(*authClientPkg.Client); ok {
				authClient.Close()
			}
		case "notifications":
			if notifClient, ok := client.(*notificationsClientPkg.Client); ok {
				notifClient.Close()
			}
		}
		delete(r.clients, serviceName)
	}

	// Находим здоровый endpoint
	healthyEndpoint := r.getHealthyEndpoint(ctx, serviceName)
	if healthyEndpoint == "" {
		return fmt.Errorf("no healthy endpoints available for service %s", serviceName)
	}

	// Создаем новое соединение
	switch serviceName {
	case "auth":
		return r.initAuthClient(ctx, healthyEndpoint)
	case "notifications":
		return r.initNotificationsClient(ctx, healthyEndpoint)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

// recordSuccess записывает успешное подключение
func (r *RedisGRPCClients) recordSuccess(ctx context.Context, serviceName, endpoint string) {
	healthKey := fmt.Sprintf("%s:health:%s:%s", r.config.KeyPrefix, serviceName, endpoint)

	healthInfo := ClientHealthInfo{
		ServiceName:  serviceName,
		Endpoint:     endpoint,
		IsHealthy:    true,
		LastCheck:    time.Now(),
		FailureCount: 0,
		CircuitOpen:  false,
	}

	if err := r.redisClient.Set(ctx, healthKey, healthInfo, r.config.HealthCheckTTL); err != nil {
		r.logger.Error("failed to record success", zap.Error(err))
	}
}

// recordFailure записывает неудачное подключение
func (r *RedisGRPCClients) recordFailure(ctx context.Context, serviceName, endpoint string, err error) {
	healthKey := fmt.Sprintf("%s:health:%s:%s", r.config.KeyPrefix, serviceName, endpoint)

	var healthInfo ClientHealthInfo
	if getErr := r.redisClient.Get(ctx, healthKey, &healthInfo); getErr != nil {
		healthInfo = ClientHealthInfo{
			ServiceName: serviceName,
			Endpoint:    endpoint,
		}
	}

	healthInfo.IsHealthy = false
	healthInfo.LastCheck = time.Now()
	healthInfo.LastError = err.Error()
	healthInfo.FailureCount++

	// Открываем circuit breaker при превышении лимита ошибок
	if r.config.CircuitBreaker && healthInfo.FailureCount >= r.config.MaxRetries {
		healthInfo.CircuitOpen = true
	}

	if setErr := r.redisClient.Set(ctx, healthKey, healthInfo, r.config.HealthCheckTTL); setErr != nil {
		r.logger.Error("failed to record failure", zap.Error(setErr))
	}
}

// updateClientStats обновляет статистику использования клиента
func (r *RedisGRPCClients) updateClientStats(ctx context.Context, serviceName string) {
	statsKey := fmt.Sprintf("%s:stats:%s", r.config.KeyPrefix, serviceName)

	// Увеличиваем счетчик запросов
	r.redisClient.HIncrBy(ctx, statsKey, "total_requests", 1)
	r.redisClient.HIncrBy(ctx, statsKey, "successful_requests", 1)
	r.redisClient.HSet(ctx, statsKey, "last_request_time", time.Now())
	r.redisClient.Expire(ctx, statsKey, 24*time.Hour)
}

// startHealthCheckRoutine запускает рутину проверки здоровья
func (r *RedisGRPCClients) startHealthCheckRoutine() {
	ticker := time.NewTicker(r.config.HealthCheckTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.performHealthChecks()
		}
	}
}

// performHealthChecks выполняет проверки здоровья всех клиентов
func (r *RedisGRPCClients) performHealthChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r.mu.RLock()
	clients := make(map[string]interface{})
	for k, v := range r.clients {
		clients[k] = v
	}
	r.mu.RUnlock()

	for serviceName, client := range clients {
		go func(svcName string, cli interface{}) {
			isHealthy := r.isClientHealthy(ctx, svcName, cli)
			if !isHealthy {
				r.logger.Warn("client health check failed", zap.String("service", svcName))
				// Записываем неудачную проверку
				r.recordFailure(ctx, svcName, "current", fmt.Errorf("health check failed"))
			}
		}(serviceName, client)
	}
}

// startStatsCollectionRoutine запускает рутину сбора статистики
func (r *RedisGRPCClients) startStatsCollectionRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.collectStats()
		}
	}
}

// collectStats собирает статистику использования
func (r *RedisGRPCClients) collectStats() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r.mu.RLock()
	services := make([]string, 0, len(r.clients))
	for serviceName := range r.clients {
		services = append(services, serviceName)
	}
	r.mu.RUnlock()

	for _, serviceName := range services {
		statsKey := fmt.Sprintf("%s:stats:%s", r.config.KeyPrefix, serviceName)
		stats, err := r.redisClient.HGetAll(ctx, statsKey)
		if err != nil {
			continue
		}

		if len(stats) > 0 {
			r.logger.Debug("client stats collected",
				zap.String("service", serviceName),
				zap.Any("stats", stats))
		}
	}
}

// GetStats возвращает статистику всех клиентов
func (r *RedisGRPCClients) GetStats(ctx context.Context) (map[string]ClientStats, error) {
	r.mu.RLock()
	services := make([]string, 0, len(r.clients))
	for serviceName := range r.clients {
		services = append(services, serviceName)
	}
	r.mu.RUnlock()

	stats := make(map[string]ClientStats)
	for _, serviceName := range services {
		statsKey := fmt.Sprintf("%s:stats:%s", r.config.KeyPrefix, serviceName)
		statsData, err := r.redisClient.HGetAll(ctx, statsKey)
		if err != nil {
			continue
		}

		clientStats := ClientStats{
			ServiceName: serviceName,
		}

		if val, ok := statsData["total_requests"]; ok {
			if err := json.Unmarshal([]byte(val), &clientStats.TotalRequests); err == nil {
				// успешно распарсили
			}
		}

		if val, ok := statsData["successful_requests"]; ok {
			if err := json.Unmarshal([]byte(val), &clientStats.SuccessfulRequests); err == nil {
				// успешно распарсили
			}
		}

		stats[serviceName] = clientStats
	}

	return stats, nil
}

// GetHealthInfo возвращает информацию о здоровье всех клиентов
func (r *RedisGRPCClients) GetHealthInfo(ctx context.Context) (map[string]ClientHealthInfo, error) {
	pattern := fmt.Sprintf("%s:health:*", r.config.KeyPrefix)
	keys, err := r.redisClient.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	healthInfo := make(map[string]ClientHealthInfo)
	for _, key := range keys {
		var info ClientHealthInfo
		if err := r.redisClient.Get(ctx, key, &info); err != nil {
			continue
		}
		healthInfo[info.ServiceName] = info
	}

	return healthInfo, nil
}

// Close закрывает все соединения
func (r *RedisGRPCClients) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errors []error

	for serviceName, client := range r.clients {
		switch serviceName {
		case "auth":
			if authClient, ok := client.(*authClientPkg.Client); ok {
				if err := authClient.Close(); err != nil {
					errors = append(errors, fmt.Errorf("failed to close auth client: %w", err))
				}
			}
		case "notifications":
			if notifClient, ok := client.(*notificationsClientPkg.Client); ok {
				if err := notifClient.Close(); err != nil {
					errors = append(errors, fmt.Errorf("failed to close notifications client: %w", err))
				}
			}
		}
	}

	r.clients = make(map[string]interface{})

	if len(errors) > 0 {
		return fmt.Errorf("errors closing gRPC clients: %v", errors)
	}

	r.logger.Info("All gRPC clients closed successfully")
	return nil
}
