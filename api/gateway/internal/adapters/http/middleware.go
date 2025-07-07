package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ===== Типы и конструкторы =====

type EndpointLimits struct {
	RPS   float64
	Burst int
}

type RateLimiter struct {
	limiters       map[string]*rate.Limiter
	mu             sync.RWMutex
	defaultRPS     float64
	defaultBurst   int
	window         time.Duration
	logger         *zap.Logger
	whitelist      map[string]bool
	whitelistMu    sync.RWMutex
	endpointLimits map[string]EndpointLimits
	endpointMu     sync.RWMutex
	stats          struct {
		requests, blocked, whitelisted int64
	}
	statsMu sync.RWMutex
}

func NewRateLimiter(rps float64, burst int, window time.Duration, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		limiters:       make(map[string]*rate.Limiter),
		defaultRPS:     rps,
		defaultBurst:   burst,
		window:         window,
		logger:         logger,
		whitelist:      make(map[string]bool),
		endpointLimits: make(map[string]EndpointLimits),
	}
	rl.setDefaultEndpointLimits()
	rl.startCleanup()
	rl.startStatsLogging()
	return rl
}

// ===== Rate Limiting: публичные методы =====

func (rl *RateLimiter) AddToWhitelist(ip string) {
	rl.whitelistMu.Lock()
	defer rl.whitelistMu.Unlock()
	rl.whitelist[ip] = true
	rl.logger.Info("IP added to whitelist", zap.String("ip", ip))
}

func (rl *RateLimiter) RemoveFromWhitelist(ip string) {
	rl.whitelistMu.Lock()
	defer rl.whitelistMu.Unlock()
	delete(rl.whitelist, ip)
	rl.logger.Info("IP removed from whitelist", zap.String("ip", ip))
}

func (rl *RateLimiter) IsWhitelisted(ip string) bool {
	rl.whitelistMu.RLock()
	defer rl.whitelistMu.RUnlock()
	return rl.whitelist[ip]
}

func NewRateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if rl.IsWhitelisted(ip) {
				rl.incrementStats(false, true)
				next.ServeHTTP(w, r)
				return
			}
			endpoint := rl.getEndpointFromPath(r.URL.Path)
			limiter := rl.getLimiter(ip, endpoint)
			if !limiter.Allow() {
				rl.incrementStats(true, false)
				rl.logger.Warn("rate limit exceeded", zap.String("ip", ip), zap.String("endpoint", endpoint), zap.String("path", r.URL.Path), zap.String("method", r.Method))
				limit, burst := rl.getLimitsForEndpoint(endpoint)
				setRateLimitHeaders(w, int(limit), 0, burst)
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests, please try again later","retry_after":1}`))
				return
			}
			rl.incrementStats(false, false)
			limit, _ := rl.getLimitsForEndpoint(endpoint)
			setRateLimitHeaders(w, int(limit), int(limit-1), 0)
			next.ServeHTTP(w, r)
		})
	}
}

// ===== JWT Middleware =====

func NewAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Check signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ===== CORS Middleware =====

func NewCORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowedOrigins == "*" && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			}
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ===== Приватные хелперы =====

func (rl *RateLimiter) setDefaultEndpointLimits() {
	rl.endpointMu.Lock()
	defer rl.endpointMu.Unlock()
	rl.endpointLimits["/api/v1/auth/login"] = EndpointLimits{RPS: 5, Burst: 10}
	rl.endpointLimits["/api/v1/auth/register"] = EndpointLimits{RPS: 2, Burst: 5}
	rl.endpointLimits["/api/v1/vpn/tunnels"] = EndpointLimits{RPS: 20, Burst: 50}
	rl.endpointLimits["/api/v1/vpn/peers"] = EndpointLimits{RPS: 30, Burst: 60}
	rl.endpointLimits["/health"] = EndpointLimits{RPS: 100, Burst: 200}
	rl.endpointLimits["/api/v1/connect"] = EndpointLimits{RPS: 10, Burst: 20}
}

func (rl *RateLimiter) getLimiter(ip, endpoint string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	key := ip + ":" + endpoint
	limiter, exists := rl.limiters[key]
	if !exists {
		rps, burst := rl.getLimitsForEndpoint(endpoint)
		limiter = rate.NewLimiter(rate.Limit(rps), burst)
		rl.limiters[key] = limiter
	}
	return limiter
}

func (rl *RateLimiter) getLimitsForEndpoint(endpoint string) (float64, int) {
	rl.endpointMu.RLock()
	defer rl.endpointMu.RUnlock()
	if l, ok := rl.endpointLimits[endpoint]; ok {
		return l.RPS, l.Burst
	}
	return rl.defaultRPS, rl.defaultBurst
}

func (rl *RateLimiter) cleanupOldLimiters() {
	ticker := time.NewTicker(rl.window)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for key := range rl.limiters {
				delete(rl.limiters, key)
			}
			rl.mu.Unlock()
			rl.logger.Debug("cleaned up old rate limiters")
		}
	}()
}

func (rl *RateLimiter) startCleanup() { rl.cleanupOldLimiters() }

func (rl *RateLimiter) startStatsLogging() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			rl.statsMu.RLock()
			requests, blocked, whitelisted := rl.stats.requests, rl.stats.blocked, rl.stats.whitelisted
			rl.statsMu.RUnlock()
			if requests > 0 {
				blockRate := float64(blocked) / float64(requests) * 100
				rl.logger.Info("rate limiting statistics", zap.Int64("total_requests", requests), zap.Int64("blocked_requests", blocked), zap.Int64("whitelisted_requests", whitelisted), zap.Float64("block_rate_percent", blockRate))
			}
		}
	}()
}

func (rl *RateLimiter) incrementStats(blocked, whitelisted bool) {
	rl.statsMu.Lock()
	defer rl.statsMu.Unlock()
	rl.stats.requests++
	if blocked {
		rl.stats.blocked++
	}
	if whitelisted {
		rl.stats.whitelisted++
	}
}

func (rl *RateLimiter) getEndpointFromPath(path string) string {
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	path = strings.TrimSuffix(path, "/")
	segments := strings.Split(path, "/")
	if len(segments) >= 3 {
		return "/" + segments[1] + "/" + segments[2]
	}
	return path
}

func setRateLimitHeaders(w http.ResponseWriter, limit, remaining, burst int) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	if burst > 0 {
		w.Header().Set("X-RateLimit-Burst", strconv.Itoa(burst))
	}
	w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))
	w.Header().Set("Content-Type", "application/json")
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if r.RemoteAddr != "" {
		if strings.Contains(r.RemoteAddr, ":") {
			return strings.Split(r.RemoteAddr, ":")[0]
		}
		return r.RemoteAddr
	}
	return "unknown"
}
