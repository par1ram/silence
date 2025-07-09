package redis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// RateLimiterAdapter адаптер для rate limiting с использованием Redis
type RateLimiterAdapter struct {
	redisClient *sharedRedis.Client
	rateLimiter *sharedRedis.RateLimiter
	logger      *zap.Logger
	config      *RateLimiterConfig
}

// RateLimiterConfig конфигурация rate limiter
type RateLimiterConfig struct {
	DefaultRPS      float64
	DefaultBurst    int
	Window          time.Duration
	KeyPrefix       string
	CleanupInterval time.Duration
	EndpointLimits  map[string]EndpointLimits
}

// EndpointLimits лимиты для конкретного endpoint
type EndpointLimits struct {
	RPS   float64
	Burst int
}

// NewRateLimiterAdapter создает новый Redis-based rate limiter adapter
func NewRateLimiterAdapter(redisClient *sharedRedis.Client, config *RateLimiterConfig, logger *zap.Logger) *RateLimiterAdapter {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "gateway:rate_limit"
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 10 * time.Minute
	}
	if config.Window == 0 {
		config.Window = time.Minute
	}
	if config.DefaultRPS == 0 {
		config.DefaultRPS = 100
	}
	if config.DefaultBurst == 0 {
		config.DefaultBurst = 200
	}
	if config.EndpointLimits == nil {
		config.EndpointLimits = make(map[string]EndpointLimits)
	}

	rateLimiter := sharedRedis.NewRateLimiter(redisClient, logger)

	adapter := &RateLimiterAdapter{
		redisClient: redisClient,
		rateLimiter: rateLimiter,
		logger:      logger,
		config:      config,
	}

	// Запускаем периодическую очистку
	go adapter.startCleanupRoutine()

	return adapter
}

// Allow проверяет, разрешен ли запрос для клиента
func (r *RateLimiterAdapter) Allow(clientIP string, endpoint string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Получаем лимиты для endpoint
	limits := r.getEndpointLimits(endpoint)

	config := &sharedRedis.RateLimitConfig{
		RPS:       limits.RPS,
		Burst:     limits.Burst,
		Window:    r.config.Window,
		KeyPrefix: r.config.KeyPrefix,
	}

	result, err := r.rateLimiter.IsAllowed(ctx, clientIP, config)
	if err != nil {
		r.logger.Error("rate limiter check failed",
			zap.String("client_ip", clientIP),
			zap.String("endpoint", endpoint),
			zap.Error(err))
		// В случае ошибки разрешаем запрос (fail open)
		return true
	}

	return result.Allowed
}

// CheckLimit проверяет лимит и возвращает детальную информацию
func (r *RateLimiterAdapter) CheckLimit(clientIP string, endpoint string) (*sharedRedis.RateLimitResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Получаем лимиты для endpoint
	limits := r.getEndpointLimits(endpoint)

	config := &sharedRedis.RateLimitConfig{
		RPS:       limits.RPS,
		Burst:     limits.Burst,
		Window:    r.config.Window,
		KeyPrefix: r.config.KeyPrefix,
	}

	result, err := r.rateLimiter.IsAllowed(ctx, clientIP, config)
	if err != nil {
		r.logger.Error("rate limiter check failed",
			zap.String("client_ip", clientIP),
			zap.String("endpoint", endpoint),
			zap.Error(err))
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	return result, nil
}

// AddToWhitelist добавляет IP в whitelist
func (r *RateLimiterAdapter) AddToWhitelist(clientIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
	}

	return r.rateLimiter.AddToWhitelist(ctx, clientIP, config)
}

// RemoveFromWhitelist удаляет IP из whitelist
func (r *RateLimiterAdapter) RemoveFromWhitelist(clientIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
	}

	return r.rateLimiter.RemoveFromWhitelist(ctx, clientIP, config)
}

// IsWhitelisted проверяет, находится ли IP в whitelist
func (r *RateLimiterAdapter) IsWhitelisted(clientIP string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
	}

	result, err := r.rateLimiter.IsAllowed(ctx, clientIP, config)
	if err != nil {
		r.logger.Error("whitelist check failed",
			zap.String("client_ip", clientIP),
			zap.Error(err))
		return false
	}

	return result.IsWhitelisted
}

// GetStats возвращает статистику rate limiter
func (r *RateLimiterAdapter) GetStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
	}

	stats, err := r.rateLimiter.GetStats(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return map[string]interface{}{
		"total_requests":   stats.TotalRequests,
		"allowed_requests": stats.AllowedRequests,
		"blocked_requests": stats.BlockedRequests,
		"unique_clients":   stats.UniqueClients,
		"top_clients":      stats.TopClients,
	}, nil
}

// ResetStats сбрасывает статистику
func (r *RateLimiterAdapter) ResetStats() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
	}

	return r.rateLimiter.ResetStats(ctx, config)
}

// SetEndpointLimits устанавливает лимиты для endpoint
func (r *RateLimiterAdapter) SetEndpointLimits(endpoint string, rps float64, burst int) {
	r.config.EndpointLimits[endpoint] = EndpointLimits{
		RPS:   rps,
		Burst: burst,
	}
}

// getEndpointLimits получает лимиты для endpoint
func (r *RateLimiterAdapter) getEndpointLimits(endpoint string) EndpointLimits {
	if limits, exists := r.config.EndpointLimits[endpoint]; exists {
		return limits
	}

	return EndpointLimits{
		RPS:   r.config.DefaultRPS,
		Burst: r.config.DefaultBurst,
	}
}

// startCleanupRoutine запускает периодическую очистку старых записей
func (r *RateLimiterAdapter) startCleanupRoutine() {
	ticker := time.NewTicker(r.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanup()
		}
	}
}

// cleanup очищает старые записи
func (r *RateLimiterAdapter) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &sharedRedis.RateLimitConfig{
		KeyPrefix: r.config.KeyPrefix,
		Window:    r.config.Window,
	}

	if err := r.rateLimiter.CleanupExpiredEntries(ctx, config); err != nil {
		r.logger.Error("failed to cleanup expired entries", zap.Error(err))
	}
}

// Middleware возвращает HTTP middleware для rate limiting
func (r *RateLimiterAdapter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Получаем IP клиента
			clientIP := r.getClientIP(req)

			// Проверяем rate limit с детальной информацией
			result, err := r.CheckLimit(clientIP, req.URL.Path)
			if err != nil {
				r.logger.Error("rate limit check error",
					zap.String("client_ip", clientIP),
					zap.String("endpoint", req.URL.Path),
					zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				r.logger.Warn("rate limit exceeded",
					zap.String("client_ip", clientIP),
					zap.String("endpoint", req.URL.Path),
					zap.Int64("remaining", result.Remaining),
					zap.Duration("retry_after", result.RetryAfter))

				// Добавляем заголовки для клиента
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))
				if result.RetryAfter > 0 {
					w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
				}

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Добавляем информационные заголовки для успешных запросов
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))

			next.ServeHTTP(w, req)
		})
	}
}

// getClientIP получает IP клиента из запроса
func (r *RateLimiterAdapter) getClientIP(req *http.Request) string {
	// Проверяем заголовки прокси
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Используем RemoteAddr как fallback
	return req.RemoteAddr
}

// Close закрывает адаптер
func (r *RateLimiterAdapter) Close() error {
	return r.redisClient.Close()
}
