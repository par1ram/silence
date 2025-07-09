package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	sharedRedis "github.com/par1ram/silence/shared/redis"
)

// WebSocketSessionManager управляет WebSocket сессиями через Redis
type WebSocketSessionManager struct {
	redisClient *sharedRedis.Client
	logger      *zap.Logger
	config      *WebSocketSessionConfig
}

// WebSocketSessionConfig конфигурация менеджера сессий
type WebSocketSessionConfig struct {
	KeyPrefix       string
	SessionTTL      time.Duration
	CleanupInterval time.Duration
	MaxSessions     int64
}

// WebSocketSession представляет WebSocket сессию
type WebSocketSession struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	ClientIP      string                 `json:"client_ip"`
	UserAgent     string                 `json:"user_agent"`
	ConnectedAt   time.Time              `json:"connected_at"`
	LastActivity  time.Time              `json:"last_activity"`
	Authenticated bool                   `json:"authenticated"`
	Subscriptions []string               `json:"subscriptions"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SessionStats статистика сессий
type SessionStats struct {
	TotalSessions          int64            `json:"total_sessions"`
	AuthenticatedSessions  int64            `json:"authenticated_sessions"`
	ActiveSessions         int64            `json:"active_sessions"`
	SessionsByUser         map[string]int64 `json:"sessions_by_user"`
	TopSubscriptions       map[string]int64 `json:"top_subscriptions"`
	AverageSessionDuration time.Duration    `json:"average_session_duration"`
}

// NewWebSocketSessionManager создает новый менеджер сессий
func NewWebSocketSessionManager(redisClient *sharedRedis.Client, config *WebSocketSessionConfig, logger *zap.Logger) *WebSocketSessionManager {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "gateway:websocket"
	}
	if config.SessionTTL == 0 {
		config.SessionTTL = 24 * time.Hour
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 10 * time.Minute
	}
	if config.MaxSessions == 0 {
		config.MaxSessions = 10000
	}

	manager := &WebSocketSessionManager{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}

	// Запускаем периодическую очистку
	go manager.startCleanupRoutine()

	return manager
}

// CreateSession создает новую сессию
func (m *WebSocketSessionManager) CreateSession(ctx context.Context, session *WebSocketSession) error {
	if session.ID == "" {
		return fmt.Errorf("session ID is required")
	}

	session.ConnectedAt = time.Now()
	session.LastActivity = time.Now()

	// Сохраняем сессию
	sessionKey := fmt.Sprintf("%s:session:%s", m.config.KeyPrefix, session.ID)
	if err := m.redisClient.Set(ctx, sessionKey, session, m.config.SessionTTL); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Добавляем в индекс активных сессий
	activeKey := fmt.Sprintf("%s:active", m.config.KeyPrefix)
	if err := m.redisClient.SAdd(ctx, activeKey, session.ID); err != nil {
		m.logger.Error("failed to add session to active index", zap.Error(err))
	}

	// Добавляем в индекс по пользователю
	if session.UserID != "" {
		userKey := fmt.Sprintf("%s:user:%s", m.config.KeyPrefix, session.UserID)
		if err := m.redisClient.SAdd(ctx, userKey, session.ID); err != nil {
			m.logger.Error("failed to add session to user index", zap.Error(err))
		}
	}

	// Обновляем статистику
	go m.updateSessionStats(ctx, session, "created")

	m.logger.Info("websocket session created",
		zap.String("session_id", session.ID),
		zap.String("user_id", session.UserID),
		zap.String("client_ip", session.ClientIP))

	return nil
}

// GetSession получает сессию по ID
func (m *WebSocketSessionManager) GetSession(ctx context.Context, sessionID string) (*WebSocketSession, error) {
	sessionKey := fmt.Sprintf("%s:session:%s", m.config.KeyPrefix, sessionID)

	var session WebSocketSession
	if err := m.redisClient.Get(ctx, sessionKey, &session); err != nil {
		if sharedRedis.IsNotFound(err) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// UpdateSession обновляет сессию
func (m *WebSocketSessionManager) UpdateSession(ctx context.Context, session *WebSocketSession) error {
	sessionKey := fmt.Sprintf("%s:session:%s", m.config.KeyPrefix, session.ID)

	session.LastActivity = time.Now()

	if err := m.redisClient.Set(ctx, sessionKey, session, m.config.SessionTTL); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Обновляем индекс по пользователю если изменился
	if session.UserID != "" {
		userKey := fmt.Sprintf("%s:user:%s", m.config.KeyPrefix, session.UserID)
		if err := m.redisClient.SAdd(ctx, userKey, session.ID); err != nil {
			m.logger.Error("failed to update user index", zap.Error(err))
		}
	}

	return nil
}

// DeleteSession удаляет сессию
func (m *WebSocketSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	// Получаем сессию для очистки индексов
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		m.logger.Error("failed to get session for deletion", zap.Error(err))
		// Продолжаем удаление даже если не смогли получить сессию
	}

	// Удаляем основную запись
	sessionKey := fmt.Sprintf("%s:session:%s", m.config.KeyPrefix, sessionID)
	if err := m.redisClient.Delete(ctx, sessionKey); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Удаляем из индекса активных сессий
	activeKey := fmt.Sprintf("%s:active", m.config.KeyPrefix)
	if err := m.redisClient.SRem(ctx, activeKey, sessionID); err != nil {
		m.logger.Error("failed to remove session from active index", zap.Error(err))
	}

	// Удаляем из индекса по пользователю
	if session != nil && session.UserID != "" {
		userKey := fmt.Sprintf("%s:user:%s", m.config.KeyPrefix, session.UserID)
		if err := m.redisClient.SRem(ctx, userKey, sessionID); err != nil {
			m.logger.Error("failed to remove session from user index", zap.Error(err))
		}
	}

	// Обновляем статистику
	if session != nil {
		go m.updateSessionStats(ctx, session, "deleted")
	}

	m.logger.Info("websocket session deleted", zap.String("session_id", sessionID))
	return nil
}

// AuthenticateSession аутентифицирует сессию
func (m *WebSocketSessionManager) AuthenticateSession(ctx context.Context, sessionID, userID string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session for authentication: %w", err)
	}

	session.Authenticated = true
	session.UserID = userID
	session.LastActivity = time.Now()

	if err := m.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update authenticated session: %w", err)
	}

	m.logger.Info("websocket session authenticated",
		zap.String("session_id", sessionID),
		zap.String("user_id", userID))

	return nil
}

// AddSubscription добавляет подписку к сессии
func (m *WebSocketSessionManager) AddSubscription(ctx context.Context, sessionID, subscription string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session for subscription: %w", err)
	}

	// Проверяем, есть ли уже такая подписка
	for _, sub := range session.Subscriptions {
		if sub == subscription {
			return nil // Подписка уже существует
		}
	}

	session.Subscriptions = append(session.Subscriptions, subscription)
	session.LastActivity = time.Now()

	if err := m.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session with subscription: %w", err)
	}

	// Добавляем в индекс подписок
	subKey := fmt.Sprintf("%s:subscription:%s", m.config.KeyPrefix, subscription)
	if err := m.redisClient.SAdd(ctx, subKey, sessionID); err != nil {
		m.logger.Error("failed to add session to subscription index", zap.Error(err))
	}

	m.logger.Debug("subscription added to session",
		zap.String("session_id", sessionID),
		zap.String("subscription", subscription))

	return nil
}

// RemoveSubscription удаляет подписку из сессии
func (m *WebSocketSessionManager) RemoveSubscription(ctx context.Context, sessionID, subscription string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session for subscription removal: %w", err)
	}

	// Удаляем подписку из списка
	newSubscriptions := make([]string, 0, len(session.Subscriptions))
	for _, sub := range session.Subscriptions {
		if sub != subscription {
			newSubscriptions = append(newSubscriptions, sub)
		}
	}

	session.Subscriptions = newSubscriptions
	session.LastActivity = time.Now()

	if err := m.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session after subscription removal: %w", err)
	}

	// Удаляем из индекса подписок
	subKey := fmt.Sprintf("%s:subscription:%s", m.config.KeyPrefix, subscription)
	if err := m.redisClient.SRem(ctx, subKey, sessionID); err != nil {
		m.logger.Error("failed to remove session from subscription index", zap.Error(err))
	}

	m.logger.Debug("subscription removed from session",
		zap.String("session_id", sessionID),
		zap.String("subscription", subscription))

	return nil
}

// GetSessionsByUser получает все сессии пользователя
func (m *WebSocketSessionManager) GetSessionsByUser(ctx context.Context, userID string) ([]*WebSocketSession, error) {
	userKey := fmt.Sprintf("%s:user:%s", m.config.KeyPrefix, userID)

	sessionIDs, err := m.redisClient.SMembers(ctx, userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*WebSocketSession, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := m.GetSession(ctx, sessionID)
		if err != nil {
			m.logger.Error("failed to get session for user",
				zap.String("session_id", sessionID),
				zap.String("user_id", userID),
				zap.Error(err))
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetSessionsBySubscription получает все сессии с определенной подпиской
func (m *WebSocketSessionManager) GetSessionsBySubscription(ctx context.Context, subscription string) ([]*WebSocketSession, error) {
	subKey := fmt.Sprintf("%s:subscription:%s", m.config.KeyPrefix, subscription)

	sessionIDs, err := m.redisClient.SMembers(ctx, subKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription sessions: %w", err)
	}

	sessions := make([]*WebSocketSession, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := m.GetSession(ctx, sessionID)
		if err != nil {
			m.logger.Error("failed to get session for subscription",
				zap.String("session_id", sessionID),
				zap.String("subscription", subscription),
				zap.Error(err))
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetStats получает статистику сессий
func (m *WebSocketSessionManager) GetStats(ctx context.Context) (*SessionStats, error) {
	stats := &SessionStats{
		SessionsByUser:   make(map[string]int64),
		TopSubscriptions: make(map[string]int64),
	}

	// Получаем общую статистику
	statsKey := fmt.Sprintf("%s:stats", m.config.KeyPrefix)
	statsData, err := m.redisClient.HGetAll(ctx, statsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	// Парсим статистику
	if val, ok := statsData["total_sessions"]; ok {
		if parsed, err := parseInt64(val); err == nil {
			stats.TotalSessions = parsed
		}
	}

	if val, ok := statsData["authenticated_sessions"]; ok {
		if parsed, err := parseInt64(val); err == nil {
			stats.AuthenticatedSessions = parsed
		}
	}

	// Получаем количество активных сессий
	activeKey := fmt.Sprintf("%s:active", m.config.KeyPrefix)
	activeCount, err := m.redisClient.SCard(ctx, activeKey)
	if err != nil {
		m.logger.Error("failed to get active sessions count", zap.Error(err))
	} else {
		stats.ActiveSessions = activeCount
	}

	return stats, nil
}

// updateSessionStats обновляет статистику сессий
func (m *WebSocketSessionManager) updateSessionStats(ctx context.Context, session *WebSocketSession, action string) {
	statsKey := fmt.Sprintf("%s:stats", m.config.KeyPrefix)

	switch action {
	case "created":
		m.redisClient.HIncrBy(ctx, statsKey, "total_sessions", 1)
		if session.Authenticated {
			m.redisClient.HIncrBy(ctx, statsKey, "authenticated_sessions", 1)
		}
	case "deleted":
		// Можем обновить статистику длительности сессий
		if !session.ConnectedAt.IsZero() {
			duration := time.Since(session.ConnectedAt)
			m.redisClient.HIncrBy(ctx, statsKey, "total_duration_seconds", int64(duration.Seconds()))
		}
	}
}

// startCleanupRoutine запускает периодическую очистку
func (m *WebSocketSessionManager) startCleanupRoutine() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup очищает устаревшие сессии
func (m *WebSocketSessionManager) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Получаем все активные сессии
	activeKey := fmt.Sprintf("%s:active", m.config.KeyPrefix)
	sessionIDs, err := m.redisClient.SMembers(ctx, activeKey)
	if err != nil {
		m.logger.Error("failed to get active sessions for cleanup", zap.Error(err))
		return
	}

	cleaned := 0
	for _, sessionID := range sessionIDs {
		// Проверяем, существует ли сессия
		sessionKey := fmt.Sprintf("%s:session:%s", m.config.KeyPrefix, sessionID)
		exists, err := m.redisClient.Exists(ctx, sessionKey)
		if err != nil {
			m.logger.Error("failed to check session existence", zap.String("session_id", sessionID), zap.Error(err))
			continue
		}

		if !exists {
			// Удаляем из индекса активных сессий
			if err := m.redisClient.SRem(ctx, activeKey, sessionID); err != nil {
				m.logger.Error("failed to remove session from active index", zap.String("session_id", sessionID), zap.Error(err))
			}
			cleaned++
		}
	}

	if cleaned > 0 {
		m.logger.Info("cleaned up stale websocket sessions", zap.Int("cleaned", cleaned))
	}
}

// parseInt64 парсит строку в int64
func parseInt64(s string) (int64, error) {
	var result int64
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return 0, err
	}
	return result, nil
}

// Close закрывает менеджер сессий
func (m *WebSocketSessionManager) Close() error {
	return m.redisClient.Close()
}
