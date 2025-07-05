package http

import (
	"context"
	"net/http"

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

	// Создаем middleware цепочку
	var rateLimitMiddleware func(http.Handler) http.Handler
	if rateLimiter != nil {
		rateLimitMiddleware = NewRateLimitMiddleware(rateLimiter)
	} else {
		rateLimitMiddleware = func(next http.Handler) http.Handler { return next }
	}

	// Регистрируем маршруты
	mux.Handle("/health", rateLimitMiddleware(http.HandlerFunc(handlers.HealthHandler)))

	// Auth маршруты
	mux.Handle("/api/v1/auth/register", rateLimitMiddleware(http.HandlerFunc(handlers.AuthHandler)))
	mux.Handle("/api/v1/auth/login", rateLimitMiddleware(http.HandlerFunc(handlers.AuthHandler)))
	// Защищённые маршруты через middleware
	mux.Handle("/api/v1/auth/me", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AuthHandler))))
	mux.Handle("/api/v1/auth/", rateLimitMiddleware(http.HandlerFunc(handlers.AuthHandler))) // fallback для остальных auth endpoint'ов

	// VPN Core маршруты (защищенные)
	mux.Handle("/api/v1/vpn/tunnels", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/list", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/get", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/start", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/stop", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/tunnels/stats", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/add", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/get", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/list", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/peers/remove", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler))))
	mux.Handle("/api/v1/vpn/", rateLimitMiddleware(http.HandlerFunc(handlers.VPNHandler))) // fallback для остальных VPN endpoint'ов

	// DPI Bypass маршруты (защищенные)
	mux.Handle("/api/v1/dpi-bypass/bypass", rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
			return
		}
		// Для других методов — старое поведение
		NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
	})))
	mux.Handle("/api/v1/dpi-bypass/bypass/", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler))))
	mux.Handle("/api/v1/dpi-bypass/", rateLimitMiddleware(http.HandlerFunc(handlers.DPIHandler))) // fallback для остальных DPI Bypass endpoint'ов

	// Интеграция VPN + обфускация (защищенный)
	mux.Handle("/api/v1/connect", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectHandler))))

	// Rate Limiting управление (только для админов, защищено auth middleware)
	mux.Handle("/api/v1/rate-limit/whitelist/add", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitWhitelistAddHandler))))
	mux.Handle("/api/v1/rate-limit/whitelist/remove", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitWhitelistRemoveHandler))))
	mux.Handle("/api/v1/rate-limit/stats", rateLimitMiddleware(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.RateLimitStatsHandler))))

	mux.Handle("/", rateLimitMiddleware(http.HandlerFunc(handlers.RootHandler)))

	server := &http.Server{
		Addr:    port,
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
