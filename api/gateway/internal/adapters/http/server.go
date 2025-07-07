package http

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// Server HTTP сервер
type Server struct {
	server   *http.Server
	handlers *Handlers
	logger   *zap.Logger
}

// NewServer создает новый HTTP сервер
func NewServer(port string, handlers *Handlers, logger *zap.Logger, jwtSecret string, rateLimiter *RateLimiter) *Server {
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
		allowedHeaders = "Content-Type,Authorization"
	}
	corsMiddleware := NewCORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders)

	// Создаем middleware цепочку
	var rateLimitMiddleware func(http.Handler) http.Handler
	if rateLimiter != nil {
		rateLimitMiddleware = NewRateLimitMiddleware(rateLimiter)
	} else {
		rateLimitMiddleware = func(next http.Handler) http.Handler { return next }
	}

	// Оборачиваем все маршруты через CORS
	wrap := func(h http.Handler) http.Handler {
		return corsMiddleware(rateLimitMiddleware(h))
	}

	// Регистрируем маршруты
	mux.Handle("/health", wrap(http.HandlerFunc(handlers.HealthHandler)))
	mux.Handle("/api/v1/auth/register", wrap(http.HandlerFunc(handlers.AuthHandler)))
	mux.Handle("/api/v1/auth/login", wrap(http.HandlerFunc(handlers.AuthHandler)))
	mux.Handle("/api/v1/auth/me", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AuthHandler))))
	mux.Handle("/api/v1/auth/", wrap(http.HandlerFunc(handlers.AuthHandler)))
	mux.Handle("/api/v1/vpn/tunnels", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/list", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/get", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/start", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/stop", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/stats", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/add", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/get", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/list", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/remove", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/", wrap(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/dpi-bypass/bypass", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
			return
		}
		NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
	})))
	mux.Handle("/api/v1/dpi-bypass/bypass/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler))))
	mux.Handle("/api/v1/dpi-bypass/", wrap(http.HandlerFunc(handlers.DPIHandler)))
	mux.Handle("/api/v1/connect", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectHandler))))
	mux.Handle("/api/v1/rate-limit/whitelist/add", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitWhitelistAddHandler))))
	mux.Handle("/api/v1/rate-limit/whitelist/remove", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitWhitelistRemoveHandler))))
	mux.Handle("/api/v1/rate-limit/stats", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitStatsHandler))))
	mux.Handle("/api/v1/analytics/metrics/connections", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/metrics/bypass-effectiveness", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/metrics/user-activity", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/metrics/server-load", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/metrics/errors", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/dashboards", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/dashboards/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/alerts", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/alerts/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AnalyticsHandler))))
	mux.Handle("/api/v1/analytics/", wrap(http.HandlerFunc(handlers.AnalyticsHandler)))
	mux.Handle("/api/v1/server-manager/servers", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ServerManagerHandler))))
	mux.Handle("/api/v1/server-manager/servers/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ServerManagerHandler))))
	mux.Handle("/api/v1/server-manager/scaling/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ServerManagerHandler))))
	mux.Handle("/api/v1/server-manager/backups/", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ServerManagerHandler))))
	mux.Handle("/api/v1/server-manager/", wrap(http.HandlerFunc(handlers.ServerManagerHandler)))
	mux.Handle("/", wrap(http.HandlerFunc(handlers.RootHandler)))

	// Добавляем двоеточие к порту для HTTP сервера
	addr := port
	if port != "" && port[0] != ':' {
		addr = ":" + port
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		server:   server,
		handlers: handlers,
		logger:   logger,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting HTTP server", zap.String("port", s.server.Addr))
	return s.server.ListenAndServe()
}

// Stop останавливает HTTP сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	return s.server.Shutdown(ctx)
}

// Name возвращает имя сервиса
func (s *Server) Name() string {
	return "gateway-http"
}
