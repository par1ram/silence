package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	redisAdapters "github.com/par1ram/silence/api/gateway/internal/adapters/redis"
	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// RedisRateLimitMiddleware создает middleware для rate limiting на базе Redis
func NewRedisRateLimitMiddleware(rateLimiter *redisAdapters.RateLimiterAdapter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем IP клиента
			clientIP := getRedisClientIP(r)
			endpoint := getRedisEndpointFromPath(r.URL.Path)

			// Проверяем rate limit с детальной информацией
			result, err := rateLimiter.CheckLimit(clientIP, endpoint)
			if err != nil {
				// В случае ошибки Redis разрешаем запрос (fail open)
				next.ServeHTTP(w, r)
				return
			}

			// Устанавливаем заголовки rate limit
			setRedisRateLimitHeaders(w, result)

			if !result.Allowed {
				// Блокируем запрос
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RedisWebSocketSessionManager создает middleware для управления WebSocket сессиями на базе Redis
func NewRedisWebSocketSessionMiddleware(sessionManager *redisAdapters.WebSocketSessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем session manager в контекст для использования в WebSocket handler
			ctx := r.Context()
			// Здесь можно добавить sessionManager в контекст если нужно
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddleware создает middleware для аутентификации JWT токенов
func NewRedisAuthMiddleware(jwtSecret string, redisClient *sharedRedis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем наличие токена
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Убираем префикс "Bearer "
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Проверяем, не находится ли токен в blacklist
			ctx := r.Context()
			isBlacklisted, err := redisClient.SIsMember(ctx, "auth:blacklist", tokenString)
			if err != nil {
				// В случае ошибки Redis продолжаем проверку токена
			} else if isBlacklisted {
				http.Error(w, "Token is blacklisted", http.StatusUnauthorized)
				return
			}

			// Парсим и валидируем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Извлекаем claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Проверяем срок действия токена
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					http.Error(w, "Token expired", http.StatusUnauthorized)
					return
				}
			}

			// Добавляем информацию о пользователе в заголовки
			if userID, ok := claims["user_id"].(string); ok {
				r.Header.Set("X-User-ID", userID)
			}
			if email, ok := claims["email"].(string); ok {
				r.Header.Set("X-User-Email", email)
			}
			if role, ok := claims["role"].(string); ok {
				r.Header.Set("X-User-Role", role)
			}

			// Опционально: обновляем активность пользователя в Redis
			if userID, ok := claims["user_id"].(string); ok {
				userActivityKey := fmt.Sprintf("user:activity:%s", userID)
				redisClient.Set(ctx, userActivityKey, time.Now().Unix(), 24*time.Hour)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware создает middleware для обработки CORS
func NewRedisCORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Устанавливаем CORS заголовки
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Обрабатываем preflight запросы
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware создает middleware для логирования запросов с сохранением в Redis
func NewRedisLoggingMiddleware(redisClient *sharedRedis.Client, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			clientIP := getRedisClientIP(r)

			// Создаем wrapper для ResponseWriter чтобы получить status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Выполняем запрос
			next.ServeHTTP(wrapped, r)

			// Логируем запрос
			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", clientIP),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("user_agent", r.Header.Get("User-Agent")),
			)

			// Сохраняем статистику в Redis
			ctx := r.Context()
			go func() {
				// Общая статистика
				redisClient.HIncrBy(ctx, "gateway:stats", "total_requests", 1)
				redisClient.HIncrBy(ctx, "gateway:stats", fmt.Sprintf("status_%d", wrapped.statusCode), 1)

				// Статистика по endpoint
				endpoint := getRedisEndpointFromPath(r.URL.Path)
				endpointKey := fmt.Sprintf("gateway:endpoint:%s", endpoint)
				redisClient.HIncrBy(ctx, endpointKey, "requests", 1)
				redisClient.HIncrBy(ctx, endpointKey, "total_duration_ms", duration.Milliseconds())

				// Статистика по клиенту
				clientKey := fmt.Sprintf("gateway:client:%s", clientIP)
				redisClient.HIncrBy(ctx, clientKey, "requests", 1)
				redisClient.Expire(ctx, clientKey, 24*time.Hour)

				// Топ клиентов
				redisClient.ZIncrBy(ctx, "gateway:top_clients", 1, clientIP)
				redisClient.Expire(ctx, "gateway:top_clients", 24*time.Hour)
			}()
		})
	}
}

// SecurityMiddleware создает middleware для безопасности
func NewRedisSecurityMiddleware(redisClient *sharedRedis.Client, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clientIP := getRedisClientIP(r)

			// Проверяем, не находится ли IP в черном списке
			isBlacklisted, err := redisClient.SIsMember(ctx, "gateway:blacklist", clientIP)
			if err != nil {
				logger.Error("failed to check blacklist", zap.Error(err))
			} else if isBlacklisted {
				logger.Warn("blocked request from blacklisted IP", zap.String("ip", clientIP))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Проверяем подозрительные patterns
			if isSuspiciousRequest(r) {
				logger.Warn("suspicious request detected",
					zap.String("ip", clientIP),
					zap.String("path", r.URL.Path),
					zap.String("user_agent", r.Header.Get("User-Agent")))

				// Увеличиваем счетчик подозрительных запросов
				suspiciousKey := fmt.Sprintf("gateway:suspicious:%s", clientIP)
				count, err := redisClient.IncrementBy(ctx, suspiciousKey, 1)
				if err != nil {
					logger.Error("failed to increment suspicious counter", zap.Error(err))
				} else {
					redisClient.Expire(ctx, suspiciousKey, time.Hour)

					// Если слишком много подозрительных запросов, блокируем IP
					if count > 10 {
						redisClient.SAdd(ctx, "gateway:blacklist", clientIP)
						logger.Warn("IP added to blacklist due to suspicious activity", zap.String("ip", clientIP))
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			}

			// Устанавливаем security заголовки
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			next.ServeHTTP(w, r)
		})
	}
}

// Вспомогательные функции

// responseWriter обертка для ResponseWriter чтобы получить status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getRedisClientIP получает IP клиента из запроса
func getRedisClientIP(r *http.Request) string {
	// Проверяем заголовки прокси
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	if xff := r.Header.Get("X-Forwarded"); xff != "" {
		return xff
	}
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
		return xfp
	}

	// Используем RemoteAddr как fallback
	if ip := strings.Split(r.RemoteAddr, ":")[0]; ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// getRedisEndpointFromPath извлекает endpoint из пути
func getRedisEndpointFromPath(path string) string {
	// Убираем query parameters
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Определяем endpoint по пути
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "api" && parts[2] == "v1" {
		return parts[3] // возвращаем основной сервис (auth, vpn, etc.)
	}

	return "unknown"
}

// setRedisRateLimitHeaders устанавливает заголовки rate limit
func setRedisRateLimitHeaders(w http.ResponseWriter, result *sharedRedis.RateLimitResult) {
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

	if result.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
	}
}

// isSuspiciousRequest проверяет, является ли запрос подозрительным
func isSuspiciousRequest(r *http.Request) bool {
	path := strings.ToLower(r.URL.Path)
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))

	// Проверяем на распространенные атаки
	suspiciousPatterns := []string{
		"../", "..",
		"<script", "javascript:",
		"union select", "drop table",
		"exec(", "eval(",
		"wp-admin", "wp-content",
		".php", ".asp", ".jsp",
		"phpmyadmin", "adminer",
		"robots.txt", "sitemap.xml",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// Проверяем подозрительные User-Agent
	if userAgent == "" ||
		strings.Contains(userAgent, "bot") ||
		strings.Contains(userAgent, "crawler") ||
		strings.Contains(userAgent, "spider") ||
		strings.Contains(userAgent, "scan") {
		return true
	}

	// Проверяем длину URL
	if len(r.URL.Path) > 1000 {
		return true
	}

	return false
}
