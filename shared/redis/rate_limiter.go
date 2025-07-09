package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter представляет Redis-based rate limiter
type RateLimiter struct {
	client *Client
	logger *zap.Logger
}

// RateLimitConfig конфигурация rate limiter
type RateLimitConfig struct {
	RPS       float64
	Burst     int
	Window    time.Duration
	KeyPrefix string
	Whitelist []string
	Blacklist []string
}

// RateLimitResult результат проверки rate limit
type RateLimitResult struct {
	Allowed       bool
	Remaining     int64
	ResetTime     time.Time
	RetryAfter    time.Duration
	IsWhitelisted bool
	IsBlacklisted bool
}

// RateLimitStats статистика rate limiter
type RateLimitStats struct {
	TotalRequests   int64
	AllowedRequests int64
	BlockedRequests int64
	WhitelistedIPs  int64
	BlacklistedIPs  int64
	UniqueClients   int64
	TopClients      map[string]int64
}

// NewRateLimiter создает новый Redis-based rate limiter
func NewRateLimiter(client *Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		client: client,
		logger: logger,
	}
}

// IsAllowed проверяет, разрешен ли запрос для клиента
func (r *RateLimiter) IsAllowed(ctx context.Context, clientIP string, config *RateLimitConfig) (*RateLimitResult, error) {
	// Проверяем whitelist
	isWhitelisted, err := r.isWhitelisted(ctx, clientIP, config)
	if err != nil {
		r.logger.Error("failed to check whitelist", zap.Error(err))
	}
	if isWhitelisted {
		return &RateLimitResult{
			Allowed:       true,
			Remaining:     int64(config.Burst),
			ResetTime:     time.Now().Add(config.Window),
			IsWhitelisted: true,
		}, nil
	}

	// Проверяем blacklist
	isBlacklisted, err := r.isBlacklisted(ctx, clientIP, config)
	if err != nil {
		r.logger.Error("failed to check blacklist", zap.Error(err))
	}
	if isBlacklisted {
		return &RateLimitResult{
			Allowed:       false,
			Remaining:     0,
			ResetTime:     time.Now().Add(config.Window),
			IsBlacklisted: true,
		}, nil
	}

	// Выполняем основную проверку rate limit
	result, err := r.checkRateLimit(ctx, clientIP, config)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return result, nil
}

// checkRateLimit выполняет основную проверку rate limit с использованием sliding window
func (r *RateLimiter) checkRateLimit(ctx context.Context, clientIP string, config *RateLimitConfig) (*RateLimitResult, error) {
	key := fmt.Sprintf("%s:rate_limit:%s", config.KeyPrefix, clientIP)
	now := time.Now()
	windowStart := now.Add(-config.Window)

	// Lua скрипт для атомарного выполнения операций
	luaScript := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local burst = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])

		-- Удаляем старые записи
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- Получаем текущее количество запросов
		local current_count = redis.call('ZCARD', key)

		-- Проверяем, не превышен ли лимит
		if current_count >= burst then
			-- Получаем время следующего сброса
			local oldest_request = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset_time = now
			if #oldest_request > 0 then
				reset_time = tonumber(oldest_request[2]) + window_ms
			end

			redis.call('EXPIRE', key, math.ceil(window_ms / 1000))
			return {0, current_count, reset_time}
		end

		-- Добавляем новый запрос
		redis.call('ZADD', key, now, now)
		redis.call('EXPIRE', key, math.ceil(window_ms / 1000))

		-- Возвращаем результат
		return {1, current_count + 1, now + window_ms}
	`

	windowStartMs := windowStart.UnixNano() / int64(time.Millisecond)
	nowMs := now.UnixNano() / int64(time.Millisecond)
	windowMs := config.Window.Nanoseconds() / int64(time.Millisecond)

	result, err := r.client.rdb.Eval(ctx, luaScript, []string{key}, windowStartMs, nowMs, config.Burst, windowMs).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit script: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 3 {
		return nil, fmt.Errorf("unexpected script result format")
	}

	allowed, _ := resultSlice[0].(int64)
	count, _ := resultSlice[1].(int64)
	resetTimeMs, _ := resultSlice[2].(int64)

	resetTime := time.Unix(0, resetTimeMs*int64(time.Millisecond))
	remaining := int64(config.Burst) - count

	var retryAfter time.Duration
	if allowed == 0 {
		retryAfter = time.Until(resetTime)
	}

	// Обновляем статистику
	go r.updateStats(ctx, clientIP, config, allowed == 1)

	return &RateLimitResult{
		Allowed:    allowed == 1,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}, nil
}

// isWhitelisted проверяет, находится ли IP в whitelist
func (r *RateLimiter) isWhitelisted(ctx context.Context, clientIP string, config *RateLimitConfig) (bool, error) {
	key := fmt.Sprintf("%s:whitelist", config.KeyPrefix)
	return r.client.SIsMember(ctx, key, clientIP)
}

// isBlacklisted проверяет, находится ли IP в blacklist
func (r *RateLimiter) isBlacklisted(ctx context.Context, clientIP string, config *RateLimitConfig) (bool, error) {
	key := fmt.Sprintf("%s:blacklist", config.KeyPrefix)
	return r.client.SIsMember(ctx, key, clientIP)
}

// AddToWhitelist добавляет IP в whitelist
func (r *RateLimiter) AddToWhitelist(ctx context.Context, clientIP string, config *RateLimitConfig) error {
	key := fmt.Sprintf("%s:whitelist", config.KeyPrefix)
	return r.client.SAdd(ctx, key, clientIP)
}

// RemoveFromWhitelist удаляет IP из whitelist
func (r *RateLimiter) RemoveFromWhitelist(ctx context.Context, clientIP string, config *RateLimitConfig) error {
	key := fmt.Sprintf("%s:whitelist", config.KeyPrefix)
	return r.client.SRem(ctx, key, clientIP)
}

// AddToBlacklist добавляет IP в blacklist
func (r *RateLimiter) AddToBlacklist(ctx context.Context, clientIP string, config *RateLimitConfig) error {
	key := fmt.Sprintf("%s:blacklist", config.KeyPrefix)
	return r.client.SAdd(ctx, key, clientIP)
}

// RemoveFromBlacklist удаляет IP из blacklist
func (r *RateLimiter) RemoveFromBlacklist(ctx context.Context, clientIP string, config *RateLimitConfig) error {
	key := fmt.Sprintf("%s:blacklist", config.KeyPrefix)
	return r.client.SRem(ctx, key, clientIP)
}

// GetWhitelist получает все IP из whitelist
func (r *RateLimiter) GetWhitelist(ctx context.Context, config *RateLimitConfig) ([]string, error) {
	key := fmt.Sprintf("%s:whitelist", config.KeyPrefix)
	return r.client.SMembers(ctx, key)
}

// GetBlacklist получает все IP из blacklist
func (r *RateLimiter) GetBlacklist(ctx context.Context, config *RateLimitConfig) ([]string, error) {
	key := fmt.Sprintf("%s:blacklist", config.KeyPrefix)
	return r.client.SMembers(ctx, key)
}

// updateStats обновляет статистику rate limiter
func (r *RateLimiter) updateStats(ctx context.Context, clientIP string, config *RateLimitConfig, allowed bool) {
	statsKey := fmt.Sprintf("%s:stats", config.KeyPrefix)
	clientKey := fmt.Sprintf("%s:client:%s", config.KeyPrefix, clientIP)

	// Обновляем общую статистику
	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, statsKey, "total_requests", 1)

	if allowed {
		pipe.HIncrBy(ctx, statsKey, "allowed_requests", 1)
	} else {
		pipe.HIncrBy(ctx, statsKey, "blocked_requests", 1)
	}

	// Обновляем статистику по клиенту
	pipe.HIncrBy(ctx, clientKey, "requests", 1)
	pipe.Expire(ctx, clientKey, 24*time.Hour) // Храним статистику клиента 24 часа

	// Добавляем в топ клиентов
	pipe.ZIncrBy(ctx, fmt.Sprintf("%s:top_clients", config.KeyPrefix), 1, clientIP)
	pipe.Expire(ctx, fmt.Sprintf("%s:top_clients", config.KeyPrefix), 24*time.Hour)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("failed to update rate limiter stats", zap.Error(err))
	}
}

// GetStats получает статистику rate limiter
func (r *RateLimiter) GetStats(ctx context.Context, config *RateLimitConfig) (*RateLimitStats, error) {
	statsKey := fmt.Sprintf("%s:stats", config.KeyPrefix)
	topClientsKey := fmt.Sprintf("%s:top_clients", config.KeyPrefix)

	// Получаем общую статистику
	statsData, err := r.client.HGetAll(ctx, statsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Получаем топ клиентов
	topClients, err := r.client.rdb.ZRevRangeWithScores(ctx, r.client.key(topClientsKey), 0, 9).Result()
	if err != nil {
		r.logger.Error("failed to get top clients", zap.Error(err))
		topClients = []redis.Z{}
	}

	// Парсим статистику
	stats := &RateLimitStats{
		TopClients: make(map[string]int64),
	}

	if val, ok := statsData["total_requests"]; ok {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			stats.TotalRequests = parsed
		}
	}

	if val, ok := statsData["allowed_requests"]; ok {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			stats.AllowedRequests = parsed
		}
	}

	if val, ok := statsData["blocked_requests"]; ok {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			stats.BlockedRequests = parsed
		}
	}

	// Заполняем топ клиентов
	for _, client := range topClients {
		if clientIP, ok := client.Member.(string); ok {
			stats.TopClients[clientIP] = int64(client.Score)
		}
	}

	// Получаем количество уникальных клиентов
	uniqueClients, err := r.client.rdb.ZCard(ctx, r.client.key(topClientsKey)).Result()
	if err != nil {
		r.logger.Error("failed to get unique clients count", zap.Error(err))
	} else {
		stats.UniqueClients = uniqueClients
	}

	return stats, nil
}

// ResetStats сбрасывает статистику rate limiter
func (r *RateLimiter) ResetStats(ctx context.Context, config *RateLimitConfig) error {
	pattern := fmt.Sprintf("%s:*", config.KeyPrefix)
	keys, err := r.client.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for reset: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Удаляем все ключи статистики
	for _, key := range keys {
		if err := r.client.Delete(ctx, key); err != nil {
			r.logger.Error("failed to delete key during reset", zap.String("key", key), zap.Error(err))
		}
	}

	r.logger.Info("rate limiter stats reset", zap.Int("deleted_keys", len(keys)))
	return nil
}

// CleanupExpiredEntries очищает истекшие записи
func (r *RateLimiter) CleanupExpiredEntries(ctx context.Context, config *RateLimitConfig) error {
	pattern := fmt.Sprintf("%s:rate_limit:*", config.KeyPrefix)
	keys, err := r.client.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for cleanup: %w", err)
	}

	now := time.Now()
	windowStart := now.Add(-config.Window)
	windowStartMs := windowStart.UnixNano() / int64(time.Millisecond)

	cleaned := 0
	for _, key := range keys {
		// Удаляем старые записи из sorted set
		removed, err := r.client.rdb.ZRemRangeByScore(ctx, r.client.key(key), "-inf", fmt.Sprintf("%d", windowStartMs)).Result()
		if err != nil {
			r.logger.Error("failed to cleanup expired entries", zap.String("key", key), zap.Error(err))
			continue
		}

		if removed > 0 {
			cleaned += int(removed)
		}

		// Если sorted set пуст, удаляем ключ
		count, err := r.client.rdb.ZCard(ctx, r.client.key(key)).Result()
		if err != nil {
			continue
		}

		if count == 0 {
			if err := r.client.Delete(ctx, key); err != nil {
				r.logger.Error("failed to delete empty key", zap.String("key", key), zap.Error(err))
			}
		}
	}

	if cleaned > 0 {
		r.logger.Info("cleaned up expired rate limit entries", zap.Int("cleaned", cleaned))
	}

	return nil
}

// GetClientStats получает статистику для конкретного клиента
func (r *RateLimiter) GetClientStats(ctx context.Context, clientIP string, config *RateLimitConfig) (map[string]string, error) {
	clientKey := fmt.Sprintf("%s:client:%s", config.KeyPrefix, clientIP)
	return r.client.HGetAll(ctx, clientKey)
}
