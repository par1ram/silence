package http

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
	logger *zap.Logger
}

func NewServer(port string, handlers *Handlers, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	// Health и root
	mux.HandleFunc("GET /health", handlers.HealthHandler)
	mux.HandleFunc("GET /", handlers.RootHandler)

	// Bypass API - используем новый синтаксис Go 1.24
	mux.HandleFunc("POST /api/v1/bypass", handlers.CreateBypass)
	mux.HandleFunc("GET /api/v1/bypass", handlers.ListBypasses)
	mux.HandleFunc("GET /api/v1/bypass/{id}", handlers.GetBypass)
	mux.HandleFunc("DELETE /api/v1/bypass/{id}", handlers.DeleteBypass)
	mux.HandleFunc("POST /api/v1/bypass/{id}/start", handlers.StartBypass)
	mux.HandleFunc("POST /api/v1/bypass/{id}/stop", handlers.StopBypass)
	mux.HandleFunc("GET /api/v1/bypass/{id}/stats", handlers.GetBypassStats)

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
		server: server,
		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting DPI Bypass HTTP server", zap.String("port", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping DPI Bypass HTTP server")
	return s.server.Shutdown(ctx)
}

func (s *Server) Name() string {
	return "dpi-bypass-http"
}
