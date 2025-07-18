package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
	"go.uber.org/zap"
)

// Handlers HTTP обработчики для auth сервиса
type Handlers struct {
	authService ports.AuthService
	logger      *zap.Logger
}

// NewHandlers создает новые HTTP обработчики
func NewHandlers(authService ports.AuthService, logger *zap.Logger) *Handlers {
	return &Handlers{
		authService: authService,
		logger:      logger,
	}
}

// RegisterHandler обработчик регистрации
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode register request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		// Проверяем тип ошибки для определения уровня логирования
		if domain.IsUserAlreadyExists(err) {
			h.logger.Warn("attempt to register existing user", zap.String("email", req.Email), zap.String("error", err.Error()))
		} else {
			h.logger.Error("failed to register user", zap.Error(err))
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode register response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LoginHandler обработчик входа
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode login request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		// Проверяем тип ошибки для определения уровня логирования
		if domain.IsInvalidCredentials(err) {
			h.logger.Warn("invalid login attempt", zap.String("email", req.Email), zap.String("error", err.Error()))
		} else {
			h.logger.Error("failed to login user", zap.Error(err))
		}
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode login response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetMeHandler обработчик получения профиля пользователя
func (h *Handlers) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Извлекаем токен (убираем "Bearer " префикс)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Валидируем токен
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.logger.Error("failed to validate token", zap.Error(err))
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Получаем профиль пользователя
	user, err := h.authService.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		h.logger.Error("failed to get user profile", zap.String("user_id", claims.UserID), zap.Error(err))
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode profile response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HealthHandler обработчик health check
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ok",
		"service": "auth",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
