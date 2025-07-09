package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	authPb "github.com/par1ram/silence/api/gateway/api/proto/auth"
	authClient "github.com/par1ram/silence/api/gateway/internal/clients/auth"
)

// GRPCProxyService сервис для проксирования запросов к gRPC сервисам
type GRPCProxyService struct {
	grpcClients *GRPCClients
	logger      *zap.Logger
}

// NewGRPCProxyService создает новый gRPC прокси сервис
func NewGRPCProxyService(grpcClients *GRPCClients, logger *zap.Logger) *GRPCProxyService {
	return &GRPCProxyService{
		grpcClients: grpcClients,
		logger:      logger,
	}
}

// ProxyAuthRequest проксирует запросы к auth сервису
func (g *GRPCProxyService) ProxyAuthRequest(w http.ResponseWriter, r *http.Request) {
	if !g.grpcClients.IsReady() {
		http.Error(w, "gRPC clients not ready", http.StatusServiceUnavailable)
		return
	}

	authClient := g.grpcClients.GetAuth()
	if authClient == nil {
		http.Error(w, "Auth service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Определяем тип запроса по пути
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/auth")
	ctx := r.Context()

	switch {
	case path == "/login" && r.Method == http.MethodPost:
		g.handleLogin(ctx, w, r, authClient)
	case path == "/register" && r.Method == http.MethodPost:
		g.handleRegister(ctx, w, r, authClient)
	case path == "/me" && r.Method == http.MethodGet:
		g.handleGetMe(ctx, w, r, authClient)
	case strings.HasPrefix(path, "/users"):
		g.handleUserManagement(ctx, w, r, authClient, path)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleLogin обрабатывает запрос на вход
func (g *GRPCProxyService) handleLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := authClient.Login(ctx, req.Email, req.Password)
	if err != nil {
		g.logger.Error("login failed", zap.Error(err))
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleRegister обрабатывает запрос на регистрацию
func (g *GRPCProxyService) handleRegister(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := authClient.Register(ctx, req.Email, req.Password)
	if err != nil {
		g.logger.Error("registration failed", zap.Error(err))
		http.Error(w, "Registration failed", http.StatusBadRequest)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusCreated)
}

// handleGetMe обрабатывает запрос на получение информации о текущем пользователе
func (g *GRPCProxyService) handleGetMe(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем префикс "Bearer " если есть
	token = strings.TrimPrefix(token, "Bearer ")

	resp, err := authClient.GetMe(ctx, token)
	if err != nil {
		g.logger.Error("get me failed", zap.Error(err))
		http.Error(w, "Failed to get user info", http.StatusUnauthorized)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleUserManagement обрабатывает запросы управления пользователями
func (g *GRPCProxyService) handleUserManagement(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	// Проверяем авторизацию для операций с пользователями
	token := r.Header.Get("X-Internal-Token")
	if token == "" {
		http.Error(w, "Internal token required", http.StatusUnauthorized)
		return
	}

	switch {
	case path == "/users" && r.Method == http.MethodGet:
		g.handleListUsers(ctx, w, r, authClient)
	case path == "/users/create" && r.Method == http.MethodPost:
		g.handleCreateUser(ctx, w, r, authClient)
	case strings.HasPrefix(path, "/users/") && r.Method == http.MethodGet:
		g.handleGetUser(ctx, w, r, authClient, path)
	case strings.HasPrefix(path, "/users/update/") && r.Method == http.MethodPut:
		g.handleUpdateUser(ctx, w, r, authClient, path)
	case strings.HasPrefix(path, "/users/delete/") && r.Method == http.MethodDelete:
		g.handleDeleteUser(ctx, w, r, authClient, path)
	case strings.HasPrefix(path, "/users/block/") && r.Method == http.MethodPost:
		g.handleBlockUser(ctx, w, r, authClient, path)
	case strings.HasPrefix(path, "/users/unblock/") && r.Method == http.MethodPost:
		g.handleUnblockUser(ctx, w, r, authClient, path)
	case strings.HasPrefix(path, "/users/role/") && r.Method == http.MethodPost:
		g.handleChangeUserRole(ctx, w, r, authClient, path)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleListUsers обрабатывает запрос на получение списка пользователей
func (g *GRPCProxyService) handleListUsers(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client) {
	// Парсим query параметры
	limit := g.getQueryParamInt32(r, "limit", 10)
	offset := g.getQueryParamInt32(r, "offset", 0)
	role := g.convertStringToUserRole(r.URL.Query().Get("role"))
	status := g.convertStringToUserStatus(r.URL.Query().Get("status"))
	email := r.URL.Query().Get("email")

	resp, err := authClient.ListUsers(ctx, role, status, email, limit, offset)
	if err != nil {
		g.logger.Error("list users failed", zap.Error(err))
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleCreateUser обрабатывает запрос на создание пользователя
func (g *GRPCProxyService) handleCreateUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := g.convertStringToUserRole(req.Role)
	resp, err := authClient.CreateUser(ctx, req.Email, req.Password, role)
	if err != nil {
		g.logger.Error("create user failed", zap.Error(err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusCreated)
}

// handleGetUser обрабатывает запрос на получение пользователя
func (g *GRPCProxyService) handleGetUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	resp, err := authClient.GetUser(ctx, id)
	if err != nil {
		g.logger.Error("get user failed", zap.Error(err))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleUpdateUser обрабатывает запрос на обновление пользователя
func (g *GRPCProxyService) handleUpdateUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/update/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := g.convertStringToUserRole(req.Role)
	status := g.convertStringToUserStatus(req.Status)

	resp, err := authClient.UpdateUser(ctx, id, req.Email, role, status)
	if err != nil {
		g.logger.Error("update user failed", zap.Error(err))
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleDeleteUser обрабатывает запрос на удаление пользователя
func (g *GRPCProxyService) handleDeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/delete/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	resp, err := authClient.DeleteUser(ctx, id)
	if err != nil {
		g.logger.Error("delete user failed", zap.Error(err))
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleBlockUser обрабатывает запрос на блокировку пользователя
func (g *GRPCProxyService) handleBlockUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/block/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	resp, err := authClient.BlockUser(ctx, id)
	if err != nil {
		g.logger.Error("block user failed", zap.Error(err))
		http.Error(w, "Failed to block user", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleUnblockUser обрабатывает запрос на разблокировку пользователя
func (g *GRPCProxyService) handleUnblockUser(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/unblock/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	resp, err := authClient.UnblockUser(ctx, id)
	if err != nil {
		g.logger.Error("unblock user failed", zap.Error(err))
		http.Error(w, "Failed to unblock user", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// handleChangeUserRole обрабатывает запрос на изменение роли пользователя
func (g *GRPCProxyService) handleChangeUserRole(ctx context.Context, w http.ResponseWriter, r *http.Request, authClient *authClient.Client, path string) {
	id := g.extractIDFromPath(path, "/users/role/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := g.convertStringToUserRole(req.Role)
	resp, err := authClient.ChangeUserRole(ctx, id, role)
	if err != nil {
		g.logger.Error("change user role failed", zap.Error(err))
		http.Error(w, "Failed to change user role", http.StatusInternalServerError)
		return
	}

	g.writeJSONResponse(w, resp, http.StatusOK)
}

// ProxyNotificationsRequest проксирует запросы к notifications сервису
func (g *GRPCProxyService) ProxyNotificationsRequest(w http.ResponseWriter, r *http.Request) {
	if !g.grpcClients.IsReady() {
		http.Error(w, "gRPC clients not ready", http.StatusServiceUnavailable)
		return
	}

	notificationsClient := g.grpcClients.GetNotifications()
	if notificationsClient == nil {
		http.Error(w, "Notifications service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Пример простого проксирования для notifications
	// В реальном проекте здесь была бы полная реализация
	ctx := r.Context()
	health, err := notificationsClient.Health(ctx)
	if err != nil {
		http.Error(w, "Notifications service unavailable", http.StatusServiceUnavailable)
		return
	}

	g.writeJSONResponse(w, health, http.StatusOK)
}

// Утилитарные методы

// writeJSONResponse записывает JSON ответ
func (g *GRPCProxyService) writeJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		g.logger.Error("failed to encode response", zap.Error(err))
	}
}

// extractIDFromPath извлекает ID из пути
func (g *GRPCProxyService) extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

// getQueryParamInt32 получает int32 параметр из query
func (g *GRPCProxyService) getQueryParamInt32(r *http.Request, param string, defaultValue int32) int32 {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}
	// Простая конвертация для примера
	// В реальном проекте нужна более надежная обработка
	return defaultValue
}

// convertStringToUserRole конвертирует строку в UserRole
func (g *GRPCProxyService) convertStringToUserRole(role string) authPb.UserRole {
	switch role {
	case "admin":
		return authPb.UserRole_USER_ROLE_ADMIN
	case "moderator":
		return authPb.UserRole_USER_ROLE_MODERATOR
	case "user":
		return authPb.UserRole_USER_ROLE_USER
	default:
		return authPb.UserRole_USER_ROLE_UNSPECIFIED
	}
}

// convertStringToUserStatus конвертирует строку в UserStatus
func (g *GRPCProxyService) convertStringToUserStatus(status string) authPb.UserStatus {
	switch status {
	case "active":
		return authPb.UserStatus_USER_STATUS_ACTIVE
	case "inactive":
		return authPb.UserStatus_USER_STATUS_INACTIVE
	case "blocked":
		return authPb.UserStatus_USER_STATUS_BLOCKED
	default:
		return authPb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}
