package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
	"go.uber.org/zap"
)

// UserHandlers обработчики для управления пользователями
type UserHandlers struct {
	userService ports.UserService
	logger      *zap.Logger
}

// NewUserHandlers создает новые обработчики пользователей
func NewUserHandlers(userService ports.UserService, logger *zap.Logger) *UserHandlers {
	return &UserHandlers{
		userService: userService,
		logger:      logger,
	}
}

// CreateUserHandler создает нового пользователя
func (h *UserHandlers) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode create user request", zap.Error(err))
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = domain.RoleUser // По умолчанию обычный пользователь
	}

	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode create user response", zap.Error(err))
	}
}

// GetUserHandler получает пользователя по ID
func (h *UserHandlers) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode get user response", zap.Error(err))
	}
}

// UpdateUserHandler обновляет пользователя
func (h *UserHandlers) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode update user request", zap.Error(err))
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update user", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode update user response", zap.Error(err))
	}
}

// DeleteUserHandler удаляет пользователя
func (h *UserHandlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	err := h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete user", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Пользователь успешно удален"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode delete user response", zap.Error(err))
	}
}

// ListUsersHandler получает список пользователей
func (h *UserHandlers) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Парсим параметры запроса
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	email := r.URL.Query().Get("email")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Создаем фильтр
	filter := &domain.UserFilter{
		Limit:  limit,
		Offset: offset,
	}

	if role != "" {
		userRole := domain.UserRole(role)
		filter.Role = &userRole
	}

	if status != "" {
		userStatus := domain.UserStatus(status)
		filter.Status = &userStatus
	}

	if email != "" {
		filter.Email = &email
	}

	response, err := h.userService.ListUsers(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode list users response", zap.Error(err))
	}
}

// BlockUserHandler блокирует пользователя
func (h *UserHandlers) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	err := h.userService.BlockUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to block user", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Пользователь заблокирован"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode block user response", zap.Error(err))
	}
}

// UnblockUserHandler разблокирует пользователя
func (h *UserHandlers) UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	err := h.userService.UnblockUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to unblock user", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Пользователь разблокирован"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode unblock user response", zap.Error(err))
	}
}

// ChangeUserRoleHandler изменяет роль пользователя
func (h *UserHandlers) ChangeUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "ID пользователя обязателен", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	var req struct {
		Role domain.UserRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode change role request", zap.Error(err))
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		http.Error(w, "Роль обязательна", http.StatusBadRequest)
		return
	}

	err := h.userService.ChangeUserRole(r.Context(), id, req.Role)
	if err != nil {
		h.logger.Error("failed to change user role", zap.String("id", id), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Роль пользователя изменена"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode change role response", zap.Error(err))
	}
}
