package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	redisAdapters "github.com/par1ram/silence/api/gateway/internal/adapters/redis"
	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// RedisServer представляет HTTP сервер с Redis-based состоянием
type RedisServer struct {
	server              *http.Server
	handlers            *Handlers
	logger              *zap.Logger
	redisClient         *sharedRedis.Client
	rateLimiter         *redisAdapters.RateLimiterAdapter
	websocketSessionMgr *redisAdapters.WebSocketSessionManager
	jwtSecret           string
}

// RedisServerConfig конфигурация Redis сервера
type RedisServerConfig struct {
	Port            string
	JWTSecret       string
	RedisHost       string
	RedisPort       int
	RedisPassword   string
	RedisDB         int
	RateLimitRPS    float64
	RateLimitBurst  int
	RateLimitWindow time.Duration
	SessionTTL      time.Duration
	CleanupInterval time.Duration
}

// NewRedisServer создает новый HTTP сервер с Redis-based состоянием
func NewRedisServer(config *RedisServerConfig, handlers *Handlers, logger *zap.Logger) (*RedisServer, error) {
	// Настройки по умолчанию
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.RedisHost == "" {
		config.RedisHost = "localhost"
	}
	if config.RedisPort == 0 {
		config.RedisPort = 6379
	}
	if config.RateLimitRPS == 0 {
		config.RateLimitRPS = 100
	}
	if config.RateLimitBurst == 0 {
		config.RateLimitBurst = 200
	}
	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = time.Minute
	}
	if config.SessionTTL == 0 {
		config.SessionTTL = 24 * time.Hour
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 10 * time.Minute
	}

	// Создаем Redis клиент
	redisConfig := &sharedRedis.Config{
		Host:     config.RedisHost,
		Port:     config.RedisPort,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
		Prefix:   "gateway",
	}

	redisClient, err := sharedRedis.NewClient(redisConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Создаем rate limiter
	rateLimiterConfig := &redisAdapters.RateLimiterConfig{
		DefaultRPS:      config.RateLimitRPS,
		DefaultBurst:    config.RateLimitBurst,
		Window:          config.RateLimitWindow,
		KeyPrefix:       "gateway:rate_limit",
		CleanupInterval: config.CleanupInterval,
		EndpointLimits: map[string]redisAdapters.EndpointLimits{
			"auth": {
				RPS:   50,
				Burst: 100,
			},
			"vpn": {
				RPS:   200,
				Burst: 400,
			},
			"analytics": {
				RPS:   30,
				Burst: 60,
			},
			"notifications": {
				RPS:   100,
				Burst: 200,
			},
		},
	}

	rateLimiter := redisAdapters.NewRateLimiterAdapter(redisClient, rateLimiterConfig, logger)

	// Создаем WebSocket session manager
	sessionConfig := &redisAdapters.WebSocketSessionConfig{
		KeyPrefix:       "gateway:websocket",
		SessionTTL:      config.SessionTTL,
		CleanupInterval: config.CleanupInterval,
		MaxSessions:     10000,
	}

	websocketSessionMgr := redisAdapters.NewWebSocketSessionManager(redisClient, sessionConfig, logger)

	// Создаем HTTP мультиплексор
	mux := http.NewServeMux()

	// Получаем CORS параметры из env
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
	if allowedMethods == "" {
		allowedMethods = "GET,POST,PUT,DELETE,OPTIONS"
	}
	allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	if allowedHeaders == "" {
		allowedHeaders = "Content-Type,Authorization,X-Requested-With"
	}

	// Создаем middleware chain
	corsMiddleware := NewRedisCORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders)
	rateLimitMiddleware := NewRedisRateLimitMiddleware(rateLimiter)
	loggingMiddleware := NewRedisLoggingMiddleware(redisClient, logger)
	securityMiddleware := NewRedisSecurityMiddleware(redisClient, logger)
	authMiddleware := NewRedisAuthMiddleware(config.JWTSecret, redisClient)

	// Функция для создания middleware chain
	chain := func(handler http.Handler, requireAuth bool) http.Handler {
		// Применяем middleware в обратном порядке
		h := handler
		if requireAuth {
			h = authMiddleware(h)
		}
		h = securityMiddleware(h)
		h = rateLimitMiddleware(h)
		h = loggingMiddleware(h)
		h = corsMiddleware(h)
		return h
	}

	// Регистрируем публичные маршруты (без аутентификации)
	mux.Handle("/health", chain(http.HandlerFunc(handlers.HealthHandler), false))
	mux.Handle("/api/v1/auth/login", chain(http.HandlerFunc(handlers.AuthHandler), false))
	mux.Handle("/api/v1/auth/register", chain(http.HandlerFunc(handlers.AuthHandler), false))

	// Регистрируем защищенные маршруты (с аутентификацией)
	mux.Handle("/api/v1/auth/me", chain(http.HandlerFunc(handlers.AuthHandler), true))
	mux.Handle("/api/v1/auth/users", chain(http.HandlerFunc(handlers.AuthHandler), true))
	mux.Handle("/api/v1/auth/users/", chain(http.HandlerFunc(handlers.AuthHandler), true))

	// VPN маршруты
	mux.Handle("/api/v1/vpn/tunnels", chain(http.HandlerFunc(handlers.VPNHandler), true))
	mux.Handle("/api/v1/vpn/tunnels/", chain(http.HandlerFunc(handlers.VPNHandler), true))
	mux.Handle("/api/v1/vpn/peers/", chain(http.HandlerFunc(handlers.VPNHandler), true))
	mux.Handle("/api/v1/vpn/", chain(http.HandlerFunc(handlers.VPNHandler), true))

	// DPI Bypass маршруты
	mux.Handle("/api/v1/dpi-bypass/", chain(http.HandlerFunc(handlers.DPIHandler), true))

	// Connection маршруты
	mux.Handle("/api/v1/connect", chain(http.HandlerFunc(handlers.ConnectHandler), true))
	mux.Handle("/api/v1/connect/", chain(http.HandlerFunc(handlers.ConnectHandler), true))
	mux.Handle("/api/v1/disconnect", chain(http.HandlerFunc(handlers.DisconnectHandler), true))

	// Analytics маршруты
	mux.Handle("/api/v1/analytics/", chain(http.HandlerFunc(handlers.AnalyticsHandler), true))

	// Server Manager маршруты
	mux.Handle("/api/v1/server-manager/", chain(http.HandlerFunc(handlers.ServerManagerHandler), true))

	// Rate Limit управление
	mux.Handle("/api/v1/rate-limit/", chain(http.HandlerFunc(handlers.RateLimitStatsHandler), true))

	// WebSocket endpoint с Redis session management
	wsHandler := NewRedisWebSocketHandler(websocketSessionMgr, redisClient, logger)
	mux.Handle("/ws", chain(http.HandlerFunc(wsHandler.HandleWebSocket), false))

	// Статистика и мониторинг
	mux.Handle("/api/v1/stats", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.StatsHandler(w, r, redisClient)
	}), true))

	// Fallback handler
	mux.Handle("/", chain(http.HandlerFunc(handlers.RootHandler), false))

	// Настраиваем адрес сервера
	addr := config.Port
	if addr != "" && addr[0] != ':' {
		addr = ":" + addr
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &RedisServer{
		server:              server,
		handlers:            handlers,
		logger:              logger,
		redisClient:         redisClient,
		rateLimiter:         rateLimiter,
		websocketSessionMgr: websocketSessionMgr,
		jwtSecret:           config.JWTSecret,
	}, nil
}

// Start запускает HTTP сервер
func (s *RedisServer) Start(ctx context.Context) error {
	// Проверяем соединение с Redis
	if err := s.redisClient.Health(ctx); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	s.logger.Info("starting Redis-based HTTP server",
		zap.String("addr", s.server.Addr),
		zap.String("redis_prefix", "gateway"))

	// Запускаем сервер
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// Stop останавливает HTTP сервер
func (s *RedisServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping Redis-based HTTP server")

	// Останавливаем HTTP сервер
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown HTTP server", zap.Error(err))
		return err
	}

	// Закрываем Redis соединения
	if err := s.redisClient.Close(); err != nil {
		s.logger.Error("failed to close Redis client", zap.Error(err))
		return err
	}

	s.logger.Info("Redis-based HTTP server stopped")
	return nil
}

// Name возвращает имя сервиса
func (s *RedisServer) Name() string {
	return "gateway-redis-http"
}

// Health проверяет состояние сервера
func (s *RedisServer) Health(ctx context.Context) error {
	// Проверяем Redis
	if err := s.redisClient.Health(ctx); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// GetStats возвращает статистику сервера
func (s *RedisServer) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Получаем статистику rate limiter
	rateLimiterStats, err := s.rateLimiter.GetStats()
	if err != nil {
		s.logger.Error("failed to get rate limiter stats", zap.Error(err))
	} else {
		stats["rate_limiter"] = rateLimiterStats
	}

	// Получаем статистику WebSocket сессий
	wsStats, err := s.websocketSessionMgr.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get websocket stats", zap.Error(err))
	} else {
		stats["websocket_sessions"] = wsStats
	}

	// Получаем общую статистику из Redis
	gatewayStats, err := s.redisClient.HGetAll(ctx, "gateway:stats")
	if err != nil {
		s.logger.Error("failed to get gateway stats", zap.Error(err))
	} else {
		stats["gateway"] = gatewayStats
	}

	return stats, nil
}

// ResetStats сбрасывает статистику
func (s *RedisServer) ResetStats(ctx context.Context) error {
	// Сбрасываем статистику rate limiter
	if err := s.rateLimiter.ResetStats(); err != nil {
		s.logger.Error("failed to reset rate limiter stats", zap.Error(err))
		return err
	}

	// Сбрасываем общую статистику
	if err := s.redisClient.Delete(ctx, "gateway:stats"); err != nil {
		s.logger.Error("failed to reset gateway stats", zap.Error(err))
		return err
	}

	s.logger.Info("server stats reset")
	return nil
}

// AddToWhitelist добавляет IP в whitelist
func (s *RedisServer) AddToWhitelist(ctx context.Context, ip string) error {
	return s.rateLimiter.AddToWhitelist(ip)
}

// RemoveFromWhitelist удаляет IP из whitelist
func (s *RedisServer) RemoveFromWhitelist(ctx context.Context, ip string) error {
	return s.rateLimiter.RemoveFromWhitelist(ip)
}

// AddToBlacklist добавляет IP в blacklist
func (s *RedisServer) AddToBlacklist(ctx context.Context, ip string) error {
	return s.redisClient.SAdd(ctx, "gateway:blacklist", ip)
}

// RemoveFromBlacklist удаляет IP из blacklist
func (s *RedisServer) RemoveFromBlacklist(ctx context.Context, ip string) error {
	return s.redisClient.SRem(ctx, "gateway:blacklist", ip)
}
