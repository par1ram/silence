package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketHandler обрабатывает WebSocket соединения
type WebSocketHandler struct {
	upgrader websocket.Upgrader
	logger   *zap.Logger
	clients  map[*websocket.Conn]*Client
	mu       sync.RWMutex
}

// Client представляет WebSocket клиента
type Client struct {
	conn          *websocket.Conn
	authenticated bool
	userID        string
	subscriptions map[string]bool
}

// WebSocketMessage представляет сообщение WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewWebSocketHandler создает новый WebSocket handler
func NewWebSocketHandler(logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// В продакшене нужно проверять origin
				return true
			},
		},
		logger:  logger,
		clients: make(map[*websocket.Conn]*Client),
	}
}

// HandleWebSocket обрабатывает WebSocket подключения
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket connection", zap.Error(err))
		return
	}
	defer h.closeConnection(conn)

	client := &Client{
		conn:          conn,
		authenticated: false,
		subscriptions: make(map[string]bool),
	}

	h.mu.Lock()
	h.clients[conn] = client
	h.mu.Unlock()

	h.logger.Info("new websocket connection established")

	// Отправляем приветственное сообщение
	h.sendToClient(conn, WebSocketMessage{
		Type:      "welcome",
		Data:      map[string]string{"message": "WebSocket connected"},
		Timestamp: time.Now(),
	})

	// Обрабатываем сообщения
	for {
		var message WebSocketMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			h.logger.Debug("websocket connection closed", zap.Error(err))
			break
		}

		h.handleMessage(conn, client, message)
	}
}

// handleMessage обрабатывает входящие сообщения
func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, client *Client, message WebSocketMessage) {
	switch message.Type {
	case "auth":
		h.handleAuth(conn, client, message)
	case "subscribe":
		h.handleSubscribe(conn, client, message)
	case "unsubscribe":
		h.handleUnsubscribe(conn, client, message)
	case "ping":
		h.handlePing(conn)
	default:
		h.sendError(conn, "unknown message type")
	}
}

// handleAuth обрабатывает аутентификацию
func (h *WebSocketHandler) handleAuth(conn *websocket.Conn, client *Client, message WebSocketMessage) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid auth data")
		return
	}

	token, ok := data["token"].(string)
	if !ok {
		h.sendError(conn, "token required")
		return
	}

	// Здесь должна быть валидация JWT токена
	// Пока просто помечаем как аутентифицированного
	if token != "" {
		client.authenticated = true
		client.userID = "user_" + token[:8] // Заглушка

		h.sendToClient(conn, WebSocketMessage{
			Type:      "auth_success",
			Data:      map[string]string{"message": "authenticated successfully"},
			Timestamp: time.Now(),
		})

		h.logger.Info("websocket client authenticated", zap.String("user_id", client.userID))
	} else {
		h.sendError(conn, "invalid token")
	}
}

// handleSubscribe обрабатывает подписку на события
func (h *WebSocketHandler) handleSubscribe(conn *websocket.Conn, client *Client, message WebSocketMessage) {
	if !client.authenticated {
		h.sendError(conn, "authentication required")
		return
	}

	data, ok := message.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid subscribe data")
		return
	}

	events, ok := data["events"].([]interface{})
	if !ok {
		h.sendError(conn, "events array required")
		return
	}

	for _, event := range events {
		if eventStr, ok := event.(string); ok {
			client.subscriptions[eventStr] = true
		}
	}

	h.sendToClient(conn, WebSocketMessage{
		Type:      "subscribe_success",
		Data:      map[string]interface{}{"events": events},
		Timestamp: time.Now(),
	})

	h.logger.Info("client subscribed to events",
		zap.String("user_id", client.userID),
		zap.Int("events_count", len(events)))
}

// handleUnsubscribe обрабатывает отписку от событий
func (h *WebSocketHandler) handleUnsubscribe(conn *websocket.Conn, client *Client, message WebSocketMessage) {
	if !client.authenticated {
		h.sendError(conn, "authentication required")
		return
	}

	data, ok := message.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "invalid unsubscribe data")
		return
	}

	events, ok := data["events"].([]interface{})
	if !ok {
		h.sendError(conn, "events array required")
		return
	}

	for _, event := range events {
		if eventStr, ok := event.(string); ok {
			delete(client.subscriptions, eventStr)
		}
	}

	h.sendToClient(conn, WebSocketMessage{
		Type:      "unsubscribe_success",
		Data:      map[string]interface{}{"events": events},
		Timestamp: time.Now(),
	})
}

// handlePing обрабатывает ping сообщения
func (h *WebSocketHandler) handlePing(conn *websocket.Conn) {
	h.sendToClient(conn, WebSocketMessage{
		Type:      "pong",
		Timestamp: time.Now(),
	})
}

// BroadcastConnectionStatus отправляет статус соединения всем подписанным клиентам
func (h *WebSocketHandler) BroadcastConnectionStatus(status interface{}) {
	message := WebSocketMessage{
		Type:      "connection_status",
		Data:      status,
		Timestamp: time.Now(),
	}

	h.broadcastToSubscribers("connection_status", message)
}

// BroadcastMetrics отправляет метрики всем подписанным клиентам
func (h *WebSocketHandler) BroadcastMetrics(metrics interface{}) {
	message := WebSocketMessage{
		Type:      "metrics_update",
		Data:      metrics,
		Timestamp: time.Now(),
	}

	h.broadcastToSubscribers("metrics_update", message)
}

// BroadcastAlert отправляет алерты всем подписанным клиентам
func (h *WebSocketHandler) BroadcastAlert(alert interface{}) {
	message := WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now(),
	}

	h.broadcastToSubscribers("alert", message)
}

// broadcastToSubscribers отправляет сообщение всем подписанным клиентам
func (h *WebSocketHandler) broadcastToSubscribers(eventType string, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn, client := range h.clients {
		if client.authenticated && client.subscriptions[eventType] {
			if err := h.sendToClient(conn, message); err != nil {
				h.logger.Error("failed to send message to client", zap.Error(err))
				// Удаляем проблемного клиента
				go h.closeConnection(conn)
			}
		}
	}
}

// sendToClient отправляет сообщение конкретному клиенту
func (h *WebSocketHandler) sendToClient(conn *websocket.Conn, message WebSocketMessage) error {
	return conn.WriteJSON(message)
}

// sendError отправляет ошибку клиенту
func (h *WebSocketHandler) sendError(conn *websocket.Conn, errorMsg string) {
	message := WebSocketMessage{
		Type:      "error",
		Data:      map[string]string{"error": errorMsg},
		Timestamp: time.Now(),
	}
	h.sendToClient(conn, message)
}

// closeConnection закрывает соединение и очищает ресурсы
func (h *WebSocketHandler) closeConnection(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, exists := h.clients[conn]; exists {
		h.logger.Info("closing websocket connection", zap.String("user_id", client.userID))
		delete(h.clients, conn)
	}

	conn.Close()
}

// GetConnectedClients возвращает количество подключенных клиентов
func (h *WebSocketHandler) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
