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
func NewServer(port string, handlers *Handlers, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	// Регистрируем маршруты
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/tunnels", handlers.CreateTunnelHandler)
	mux.HandleFunc("/tunnels/list", handlers.ListTunnelsHandler)
	mux.HandleFunc("/tunnels/get", handlers.GetTunnelHandler)
	mux.HandleFunc("/tunnels/start", handlers.StartTunnelHandler)
	mux.HandleFunc("/tunnels/stop", handlers.StopTunnelHandler)
	mux.HandleFunc("/tunnels/stats", handlers.GetTunnelStatsHandler)
	mux.HandleFunc("/peers/add", handlers.AddPeerHandler)
	mux.HandleFunc("/peers/get", handlers.GetPeerHandler)
	mux.HandleFunc("/peers/list", handlers.ListPeersHandler)
	mux.HandleFunc("/peers/remove", handlers.RemovePeerHandler)

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
	s.logger.Info("starting VPN Core HTTP server", zap.String("port", s.server.Addr))
	return s.server.ListenAndServe()
}

// Stop останавливает HTTP сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping VPN Core HTTP server")
	return s.server.Shutdown(ctx)
}

// Name возвращает имя сервиса
func (s *Server) Name() string {
	return "vpn-core-http"
}
