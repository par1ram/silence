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
func NewServer(port string, handlers *Handlers, logger *zap.Logger, jwtSecret string) *Server {
	mux := http.NewServeMux()

	// Регистрируем маршруты
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Auth маршруты
	mux.HandleFunc("/api/v1/auth/register", handlers.AuthHandler)
	mux.HandleFunc("/api/v1/auth/login", handlers.AuthHandler)
	// Защищённые маршруты через middleware
	mux.Handle("/api/v1/auth/me", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.AuthHandler)))
	mux.HandleFunc("/api/v1/auth/", handlers.AuthHandler) // fallback для остальных auth endpoint'ов

	// VPN Core маршруты (защищенные)
	mux.Handle("/api/v1/vpn/tunnels", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/tunnels/list", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/tunnels/get", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/tunnels/start", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/tunnels/stop", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/tunnels/stats", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/peers/add", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/peers/get", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/peers/list", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.Handle("/api/v1/vpn/peers/remove", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.VPNHandler)))
	mux.HandleFunc("/api/v1/vpn/", handlers.VPNHandler) // fallback для остальных VPN endpoint'ов

	// DPI Bypass маршруты (защищенные)
	mux.HandleFunc("/api/v1/dpi-bypass/bypass", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
			return
		}
		// Для других методов — старое поведение
		NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)).ServeHTTP(w, r)
	})
	mux.Handle("/api/v1/dpi-bypass/bypass/", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DPIHandler)))
	mux.HandleFunc("/api/v1/dpi-bypass/", handlers.DPIHandler) // fallback для остальных DPI Bypass endpoint'ов

	// Интеграция VPN + обфускация (защищенный)
	mux.Handle("/api/v1/connect", NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectHandler)))

	mux.HandleFunc("/", handlers.RootHandler)

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
