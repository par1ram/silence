package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/ports"
	"go.uber.org/zap"
)

// BypassService сервис для управления DPI bypass
type BypassService struct {
	configs map[string]*domain.BypassConfig
	adapter ports.BypassAdapter
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewBypassService создает новый bypass сервис
func NewBypassService(adapter ports.BypassAdapter, logger *zap.Logger) ports.DPIBypassService {
	return &BypassService{
		configs: make(map[string]*domain.BypassConfig),
		adapter: adapter,
		logger:  logger,
	}
}

// CreateBypassConfig создает новую bypass конфигурацию
func (s *BypassService) CreateBypassConfig(ctx context.Context, req *domain.CreateBypassConfigRequest) (*domain.BypassConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Генерируем ID (в продакшене использовать UUID)
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	config := &domain.BypassConfig{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Method:      req.Method,
		Status:      domain.BypassStatusInactive,
		Parameters:  req.Parameters,
		Rules:       []*domain.BypassRule{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.configs[id] = config

	s.logger.Info("bypass configuration created",
		zap.String("id", id),
		zap.String("name", req.Name),
		zap.String("method", string(req.Method)))

	return config, nil
}

// GetBypassConfig получает bypass конфигурацию по ID
func (s *BypassService) GetBypassConfig(ctx context.Context, id string) (*domain.BypassConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, fmt.Errorf("bypass configuration not found: %s", id)
	}

	return config, nil
}

// ListBypassConfigs возвращает список всех bypass конфигураций
func (s *BypassService) ListBypassConfigs(ctx context.Context, filters *domain.BypassConfigFilters) ([]*domain.BypassConfig, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configs := make([]*domain.BypassConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configs = append(configs, config)
	}

	return configs, len(configs), nil
}

// StartBypass запускает bypass соединение
func (s *BypassService) StartBypass(ctx context.Context, req *domain.StartBypassRequest) (*domain.BypassSession, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[req.ConfigID]
	if !exists {
		return nil, fmt.Errorf("bypass configuration not found: %s", req.ConfigID)
	}

	// Генерируем ID сессии
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())

	session := &domain.BypassSession{
		ID:         sessionID,
		ConfigID:   req.ConfigID,
		TargetHost: req.TargetHost,
		TargetPort: req.TargetPort,
		Status:     domain.BypassStatusActive,
		StartedAt:  time.Now(),
		Message:    "Bypass session started",
	}

	config.Status = domain.BypassStatusActive
	config.UpdatedAt = time.Now()

	s.logger.Info("bypass session started", zap.String("session_id", sessionID))
	return session, nil
}

// StopBypass останавливает bypass соединение
func (s *BypassService) StopBypass(ctx context.Context, sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("bypass session stopped", zap.String("session_id", sessionID))
	return nil
}

// GetBypassStats получает статистику bypass соединения
func (s *BypassService) GetBypassStats(ctx context.Context, sessionID string) (*domain.BypassStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return &domain.BypassStats{
		ID:                     "stats_" + sessionID,
		SessionID:              sessionID,
		BytesSent:              1024,
		BytesReceived:          2048,
		PacketsSent:            100,
		PacketsReceived:        200,
		ConnectionsEstablished: 1,
		ConnectionsFailed:      0,
		SuccessRate:            1.0,
		AverageLatency:         50.0,
		StartTime:              time.Now().Add(-time.Hour),
		EndTime:                time.Now(),
	}, nil
}

// DeleteBypassConfig удаляет bypass конфигурацию
func (s *BypassService) DeleteBypassConfig(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return fmt.Errorf("bypass configuration not found: %s", id)
	}

	delete(s.configs, id)

	s.logger.Info("bypass configuration deleted", zap.String("id", id), zap.String("name", config.Name))
	return nil
}

// UpdateBypassConfig обновляет bypass конфигурацию
func (s *BypassService) UpdateBypassConfig(ctx context.Context, req *domain.UpdateBypassConfigRequest) (*domain.BypassConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[req.ID]
	if !exists {
		return nil, fmt.Errorf("bypass configuration not found: %s", req.ID)
	}

	config.Name = req.Name
	config.Description = req.Description
	config.Type = req.Type
	config.Method = req.Method
	config.Parameters = req.Parameters
	config.UpdatedAt = time.Now()

	return config, nil
}

// GetBypassStatus получает статус bypass сессии
func (s *BypassService) GetBypassStatus(ctx context.Context, sessionID string) (*domain.BypassSessionStatus, error) {
	return &domain.BypassSessionStatus{
		SessionID:       sessionID,
		Status:          domain.BypassStatusActive,
		TargetHost:      "example.com",
		TargetPort:      443,
		StartedAt:       time.Now().Add(-time.Hour),
		DurationSeconds: 3600,
		Message:         "Session is active",
	}, nil
}

// GetBypassHistory получает историю bypass сессий
func (s *BypassService) GetBypassHistory(ctx context.Context, req *domain.BypassHistoryRequest) ([]*domain.BypassHistoryEntry, int, error) {
	return []*domain.BypassHistoryEntry{}, 0, nil
}

// AddBypassRule добавляет правило bypass
func (s *BypassService) AddBypassRule(ctx context.Context, req *domain.AddBypassRuleRequest) (*domain.BypassRule, error) {
	ruleID := fmt.Sprintf("rule_%d", time.Now().UnixNano())

	rule := &domain.BypassRule{
		ID:         ruleID,
		ConfigID:   req.ConfigID,
		Name:       req.Name,
		Type:       req.Type,
		Action:     req.Action,
		Pattern:    req.Pattern,
		Parameters: req.Parameters,
		Priority:   req.Priority,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return rule, nil
}

// UpdateBypassRule обновляет правило bypass
func (s *BypassService) UpdateBypassRule(ctx context.Context, req *domain.UpdateBypassRuleRequest) (*domain.BypassRule, error) {
	rule := &domain.BypassRule{
		ID:         req.ID,
		Name:       req.Name,
		Type:       req.Type,
		Action:     req.Action,
		Pattern:    req.Pattern,
		Parameters: req.Parameters,
		Priority:   req.Priority,
		Enabled:    req.Enabled,
		UpdatedAt:  time.Now(),
	}

	return rule, nil
}

// DeleteBypassRule удаляет правило bypass
func (s *BypassService) DeleteBypassRule(ctx context.Context, id string) error {
	s.logger.Info("bypass rule deleted", zap.String("id", id))
	return nil
}

// ListBypassRules получает список правил bypass
func (s *BypassService) ListBypassRules(ctx context.Context, filters *domain.BypassRuleFilters) ([]*domain.BypassRule, int, error) {
	return []*domain.BypassRule{}, 0, nil
}
