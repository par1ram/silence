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
	if !h.validateMethod(w, r, http.MethodPost) {
		return
	}

	var req domain.CreateUserRequest
	if !h.decodeRequest(w, r, &req) {
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		h.writeError(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = domain.RoleUser // По умолчанию обычный пользователь
	}

	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, user, http.StatusCreated)
}

// GetUserHandler получает пользователя по ID
func (h *UserHandlers) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	h.writeJSON(w, user, http.StatusOK)
}

// UpdateUserHandler обновляет пользователя
func (h *UserHandlers) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodPut) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	var req domain.UpdateUserRequest
	if !h.decodeRequest(w, r, &req) {
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update user", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, user, http.StatusOK)
}

// DeleteUserHandler удаляет пользователя
func (h *UserHandlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodDelete) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	err := h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete user", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"message": "Пользователь успешно удален"}, http.StatusOK)
}

// ListUsersHandler получает список пользователей
func (h *UserHandlers) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	filter := h.buildUserFilter(r)
	response, err := h.userService.ListUsers(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, response, http.StatusOK)
}

// BlockUserHandler блокирует пользователя
func (h *UserHandlers) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodPost) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	err := h.userService.BlockUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to block user", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"message": "Пользователь заблокирован"}, http.StatusOK)
}

// UnblockUserHandler разблокирует пользователя
func (h *UserHandlers) UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodPost) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	err := h.userService.UnblockUser(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to unblock user", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"message": "Пользователь разблокирован"}, http.StatusOK)
}

// ChangeUserRoleHandler изменяет роль пользователя
func (h *UserHandlers) ChangeUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodPost) {
		return
	}

	id := h.extractID(w, r)
	if id == "" {
		return
	}

	var req struct {
		Role domain.UserRole `json:"role"`
	}
	if !h.decodeRequest(w, r, &req) {
		return
	}

	if req.Role == "" {
		h.writeError(w, "Роль обязательна", http.StatusBadRequest)
		return
	}

	err := h.userService.ChangeUserRole(r.Context(), id, req.Role)
	if err != nil {
		h.logger.Error("failed to change user role", zap.String("id", id), zap.Error(err))
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"message": "Роль пользователя изменена"}, http.StatusOK)
}

// Утилиты для обработки HTTP запросов

// validateMethod проверяет метод HTTP запроса
func (h *UserHandlers) validateMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// extractID извлекает ID из URL
func (h *UserHandlers) extractID(w http.ResponseWriter, r *http.Request) string {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.writeError(w, "ID пользователя обязателен", http.StatusBadRequest)
		return ""
	}
	return pathParts[len(pathParts)-1]
}

// decodeRequest декодирует JSON из тела запроса
func (h *UserHandlers) decodeRequest(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		h.writeError(w, "Неверный формат данных", http.StatusBadRequest)
		return false
	}
	return true
}

// writeJSON записывает JSON ответ
func (h *UserHandlers) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// writeError записывает ошибку
func (h *UserHandlers) writeError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}

// buildUserFilter создает фильтр для списка пользователей
func (h *UserHandlers) buildUserFilter(r *http.Request) *domain.UserFilter {
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

	return filter
}
