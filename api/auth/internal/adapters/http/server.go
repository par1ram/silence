package http

import (
	"context"
	"net/http"

	"github.com/par1ram/silence/api/auth/internal/config"
	"go.uber.org/zap"
)

// Server HTTP сервер для auth сервиса
type Server struct {
	server       *http.Server
	handlers     *Handlers
	userHandlers *UserHandlers
	logger       *zap.Logger
}

// NewServer создает новый HTTP сервер
func NewServer(port string, handlers *Handlers, userHandlers *UserHandlers, cfg *config.Config, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	// Регистрируем маршруты аутентификации
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Мидлварь для внутренних сервисов
	internalMW := InternalTokenMiddleware(cfg.InternalAPIToken)

	mux.Handle("/users", internalMW(http.HandlerFunc(userHandlers.ListUsersHandler)))
	mux.Handle("/users/", internalMW(http.HandlerFunc(userHandlers.GetUserHandler)))
	mux.Handle("/users/create", internalMW(http.HandlerFunc(userHandlers.CreateUserHandler)))
	mux.Handle("/users/update/", internalMW(http.HandlerFunc(userHandlers.UpdateUserHandler)))
	mux.Handle("/users/delete/", internalMW(http.HandlerFunc(userHandlers.DeleteUserHandler)))
	mux.Handle("/users/block/", internalMW(http.HandlerFunc(userHandlers.BlockUserHandler)))
	mux.Handle("/users/unblock/", internalMW(http.HandlerFunc(userHandlers.UnblockUserHandler)))
	mux.Handle("/users/role/", internalMW(http.HandlerFunc(userHandlers.ChangeUserRoleHandler)))

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	return &Server{
		server:       server,
		handlers:     handlers,
		userHandlers: userHandlers,
		logger:       logger,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting auth HTTP server", zap.String("port", s.server.Addr))
	return s.server.ListenAndServe()
}

// Stop останавливает HTTP сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping auth HTTP server")
	return s.server.Shutdown(ctx)
}

// Name возвращает имя сервиса
func (s *Server) Name() string {
	return "auth-http"
}

// GetHandler возвращает HTTP handler для тестирования
func (s *Server) GetHandler() http.Handler {
	return s.server.Handler
}
