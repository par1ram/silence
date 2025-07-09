package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	redisAdapters "github.com/par1ram/silence/api/gateway/internal/adapters/redis"
	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// RedisWebSocketHandler обрабатывает WebSocket соединения с Redis-based управлением сессиями
type RedisWebSocketHandler struct {
	sessionManager *redisAdapters.WebSocketSessionManager
	redisClient    *sharedRedis.Client
	logger         *zap.Logger
	upgrader       websocket.Upgrader
}

// RedisWebSocketMessage представляет сообщение WebSocket для Redis handler
type RedisWebSocketMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewRedisWebSocketHandler создает новый WebSocket handler с Redis поддержкой
func NewRedisWebSocketHandler(sessionManager *redisAdapters.WebSocketSessionManager, redisClient *sharedRedis.Client, logger *zap.Logger) *RedisWebSocketHandler {
	return &RedisWebSocketHandler{
		sessionManager: sessionManager,
		redisClient:    redisClient,
		logger:         logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// В продакшене здесь должна быть более строгая проверка Origin
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (h *RedisWebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Апгрейдим HTTP соединение до WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Создаем сессию
	sessionID := h.generateSessionID()
	session := &redisAdapters.WebSocketSession{
		ID:            sessionID,
		ClientIP:      h.getClientIP(r),
		UserAgent:     r.Header.Get("User-Agent"),
		ConnectedAt:   time.Now(),
		LastActivity:  time.Now(),
		Authenticated: false,
		Subscriptions: make([]string, 0),
		Metadata:      make(map[string]interface{}),
	}

	ctx := context.Background()
	if err := h.sessionManager.CreateSession(ctx, session); err != nil {
		h.logger.Error("failed to create websocket session", zap.Error(err))
		return
	}

	h.logger.Info("websocket connection established",
		zap.String("session_id", sessionID),
		zap.String("client_ip", session.ClientIP))

	// Отправляем приветственное сообщение
	welcomeMsg := RedisWebSocketMessage{
		Type:      "welcome",
		ID:        sessionID,
		Data:      map[string]interface{}{"session_id": sessionID},
		Timestamp: time.Now(),
	}

	if err := h.sendMessage(conn, welcomeMsg); err != nil {
		h.logger.Error("failed to send welcome message", zap.Error(err))
		return
	}

	// Запускаем горутины для обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Горутина для чтения сообщений от клиента
	go func() {
		defer cancel()
		h.handleIncomingMessages(ctx, conn, sessionID)
	}()

	// Горутина для отправки сообщений клиенту
	go func() {
		defer cancel()
		h.handleOutgoingMessages(ctx, conn, sessionID)
	}()

	// Горутина для ping/pong
	go func() {
		defer cancel()
		h.handlePingPong(ctx, conn, sessionID)
	}()

	// Ожидаем завершения
	<-ctx.Done()

	// Очищаем сессию
	if err := h.sessionManager.DeleteSession(context.Background(), sessionID); err != nil {
		h.logger.Error("failed to delete websocket session", zap.Error(err))
	}

	h.logger.Info("websocket connection closed",
		zap.String("session_id", sessionID))
}

// handleIncomingMessages обрабатывает входящие сообщения от клиента
func (h *RedisWebSocketHandler) handleIncomingMessages(ctx context.Context, conn *websocket.Conn, sessionID string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg RedisWebSocketMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("websocket read error", zap.Error(err))
				}
				return
			}

			// Обновляем время последней активности
			session, err := h.sessionManager.GetSession(ctx, sessionID)
			if err != nil {
				h.logger.Error("failed to get session", zap.Error(err))
				continue
			}

			session.LastActivity = time.Now()
			if err := h.sessionManager.UpdateSession(ctx, session); err != nil {
				h.logger.Error("failed to update session", zap.Error(err))
			}

			// Обрабатываем сообщение
			h.processMessage(ctx, conn, sessionID, msg)
		}
	}
}

// handleOutgoingMessages обрабатывает исходящие сообщения для клиента
func (h *RedisWebSocketHandler) handleOutgoingMessages(ctx context.Context, conn *websocket.Conn, sessionID string) {
	// Подписываемся на сообщения для данной сессии
	pubsubKey := fmt.Sprintf("websocket:session:%s:messages", sessionID)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Проверяем наличие сообщений для отправки
			messages, err := h.redisClient.LRange(ctx, pubsubKey, 0, -1)
			if err != nil {
				continue
			}

			for _, msgStr := range messages {
				var msg RedisWebSocketMessage
				if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
					h.logger.Error("failed to unmarshal message", zap.Error(err))
					continue
				}

				if err := h.sendMessage(conn, msg); err != nil {
					h.logger.Error("failed to send message", zap.Error(err))
					return
				}
			}

			// Очищаем отправленные сообщения
			if len(messages) > 0 {
				h.redisClient.Delete(ctx, pubsubKey)
			}
		}
	}
}

// handlePingPong обрабатывает ping/pong для поддержания соединения
func (h *RedisWebSocketHandler) handlePingPong(ctx context.Context, conn *websocket.Conn, sessionID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				h.logger.Error("failed to send ping", zap.Error(err))
				return
			}

			// Обновляем время последней активности
			session, err := h.sessionManager.GetSession(ctx, sessionID)
			if err != nil {
				continue
			}

			session.LastActivity = time.Now()
			h.sessionManager.UpdateSession(ctx, session)
		}
	}
}

// processMessage обрабатывает конкретное сообщение от клиента
func (h *RedisWebSocketHandler) processMessage(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	switch msg.Type {
	case "auth":
		h.handleAuth(ctx, conn, sessionID, msg)
	case "subscribe":
		h.handleSubscribe(ctx, conn, sessionID, msg)
	case "unsubscribe":
		h.handleUnsubscribe(ctx, conn, sessionID, msg)
	case "ping":
		h.handlePing(ctx, conn, sessionID, msg)
	case "message":
		h.handleMessage(ctx, conn, sessionID, msg)
	default:
		h.logger.Warn("unknown message type",
			zap.String("type", msg.Type),
			zap.String("session_id", sessionID))

		errorMsg := RedisWebSocketMessage{
			Type:      "error",
			Data:      map[string]interface{}{"error": "unknown message type"},
			Timestamp: time.Now(),
		}
		h.sendMessage(conn, errorMsg)
	}
}

// handleAuth обрабатывает аутентификацию пользователя
func (h *RedisWebSocketHandler) handleAuth(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid auth data")
		return
	}

	userID, ok := data["user_id"].(string)
	if !ok {
		h.sendError(conn, "user_id is required")
		return
	}

	if err := h.sessionManager.AuthenticateSession(ctx, sessionID, userID); err != nil {
		h.logger.Error("failed to authenticate session", zap.Error(err))
		h.sendError(conn, "authentication failed")
		return
	}

	response := RedisWebSocketMessage{
		Type:      "auth_success",
		Data:      map[string]interface{}{"user_id": userID},
		Timestamp: time.Now(),
	}

	h.sendMessage(conn, response)
}

// handleSubscribe обрабатывает подписку на события
func (h *RedisWebSocketHandler) handleSubscribe(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid subscribe data")
		return
	}

	subscription, ok := data["subscription"].(string)
	if !ok {
		h.sendError(conn, "subscription is required")
		return
	}

	if err := h.sessionManager.AddSubscription(ctx, sessionID, subscription); err != nil {
		h.logger.Error("failed to add subscription", zap.Error(err))
		h.sendError(conn, "subscription failed")
		return
	}

	response := RedisWebSocketMessage{
		Type:      "subscribed",
		Data:      map[string]interface{}{"subscription": subscription},
		Timestamp: time.Now(),
	}

	h.sendMessage(conn, response)
}

// handleUnsubscribe обрабатывает отписку от событий
func (h *RedisWebSocketHandler) handleUnsubscribe(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid unsubscribe data")
		return
	}

	subscription, ok := data["subscription"].(string)
	if !ok {
		h.sendError(conn, "subscription is required")
		return
	}

	if err := h.sessionManager.RemoveSubscription(ctx, sessionID, subscription); err != nil {
		h.logger.Error("failed to remove subscription", zap.Error(err))
		h.sendError(conn, "unsubscribe failed")
		return
	}

	response := RedisWebSocketMessage{
		Type:      "unsubscribed",
		Data:      map[string]interface{}{"subscription": subscription},
		Timestamp: time.Now(),
	}

	h.sendMessage(conn, response)
}

// handlePing обрабатывает ping от клиента
func (h *RedisWebSocketHandler) handlePing(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	response := RedisWebSocketMessage{
		Type:      "pong",
		Timestamp: time.Now(),
	}

	h.sendMessage(conn, response)
}

// handleMessage обрабатывает обычное сообщение
func (h *RedisWebSocketHandler) handleMessage(ctx context.Context, conn *websocket.Conn, sessionID string, msg RedisWebSocketMessage) {
	// Здесь можно добавить логику обработки пользовательских сообщений
	h.logger.Info("received message",
		zap.String("session_id", sessionID),
		zap.String("type", msg.Type),
		zap.Any("data", msg.Data))

	// Отправляем подтверждение
	response := RedisWebSocketMessage{
		Type:      "message_received",
		ID:        msg.ID,
		Timestamp: time.Now(),
	}

	h.sendMessage(conn, response)
}

// sendMessage отправляет сообщение клиенту
func (h *RedisWebSocketHandler) sendMessage(conn *websocket.Conn, msg RedisWebSocketMessage) error {
	return conn.WriteJSON(msg)
}

// sendError отправляет сообщение об ошибке
func (h *RedisWebSocketHandler) sendError(conn *websocket.Conn, errorMsg string) {
	msg := RedisWebSocketMessage{
		Type:      "error",
		Data:      map[string]interface{}{"error": errorMsg},
		Timestamp: time.Now(),
	}

	if err := h.sendMessage(conn, msg); err != nil {
		h.logger.Error("failed to send error message", zap.Error(err))
	}
}

// getClientIP получает IP клиента
func (h *RedisWebSocketHandler) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// generateSessionID генерирует уникальный ID сессии
func (h *RedisWebSocketHandler) generateSessionID() string {
	return fmt.Sprintf("ws_%d", time.Now().UnixNano())
}

// BroadcastToSubscription отправляет сообщение всем подписчикам
func (h *RedisWebSocketHandler) BroadcastToSubscription(ctx context.Context, subscription string, msg RedisWebSocketMessage) error {
	sessions, err := h.sessionManager.GetSessionsBySubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to get sessions for subscription: %w", err)
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, session := range sessions {
		messageKey := fmt.Sprintf("websocket:session:%s:messages", session.ID)
		if err := h.redisClient.LPush(ctx, messageKey, string(msgData)); err != nil {
			h.logger.Error("failed to queue message for session",
				zap.String("session_id", session.ID),
				zap.Error(err))
		}
	}

	return nil
}

// BroadcastToUser отправляет сообщение всем сессиям пользователя
func (h *RedisWebSocketHandler) BroadcastToUser(ctx context.Context, userID string, msg RedisWebSocketMessage) error {
	sessions, err := h.sessionManager.GetSessionsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get sessions for user: %w", err)
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, session := range sessions {
		messageKey := fmt.Sprintf("websocket:session:%s:messages", session.ID)
		if err := h.redisClient.LPush(ctx, messageKey, string(msgData)); err != nil {
			h.logger.Error("failed to queue message for session",
				zap.String("session_id", session.ID),
				zap.Error(err))
		}
	}

	return nil
}

// GetSessionStats возвращает статистику WebSocket сессий
func (h *RedisWebSocketHandler) GetSessionStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := h.sessionManager.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_sessions":           stats.TotalSessions,
		"authenticated_sessions":   stats.AuthenticatedSessions,
		"active_sessions":          stats.ActiveSessions,
		"sessions_by_user":         stats.SessionsByUser,
		"top_subscriptions":        stats.TopSubscriptions,
		"average_session_duration": stats.AverageSessionDuration.String(),
	}, nil
}
