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
func NewBypassService(adapter ports.BypassAdapter, logger *zap.Logger) *BypassService {
	return &BypassService{
		configs: make(map[string]*domain.BypassConfig),
		adapter: adapter,
		logger:  logger,
	}
}

// CreateBypass создает новую bypass конфигурацию
func (s *BypassService) CreateBypass(ctx context.Context, req *domain.CreateBypassRequest) (*domain.BypassConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Генерируем ID (в продакшене использовать UUID)
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	config := &domain.BypassConfig{
		ID:         id,
		Name:       req.Name,
		Method:     req.Method,
		LocalPort:  req.LocalPort,
		RemoteHost: req.RemoteHost,
		RemotePort: req.RemotePort,
		Password:   req.Password,
		Encryption: req.Encryption,
		Status:     "inactive",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.configs[id] = config

	s.logger.Info("bypass configuration created",
		zap.String("id", id),
		zap.String("name", req.Name),
		zap.String("method", string(req.Method)))

	return config, nil
}

// GetBypass получает bypass конфигурацию по ID
func (s *BypassService) GetBypass(ctx context.Context, id string) (*domain.BypassConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, fmt.Errorf("bypass configuration not found: %s", id)
	}

	return config, nil
}

// ListBypasses возвращает список всех bypass конфигураций
func (s *BypassService) ListBypasses(ctx context.Context) ([]*domain.BypassConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configs := make([]*domain.BypassConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configs = append(configs, config)
	}

	return configs, nil
}

// StartBypass запускает bypass соединение
func (s *BypassService) StartBypass(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return fmt.Errorf("bypass configuration not found: %s", id)
	}

	if s.adapter.IsRunning(id) {
		return fmt.Errorf("bypass connection already running: %s", id)
	}

	if err := s.adapter.Start(config); err != nil {
		s.logger.Error("failed to start bypass connection", zap.Error(err), zap.String("id", id))
		return err
	}

	config.Status = "active"
	config.UpdatedAt = time.Now()

	s.logger.Info("bypass connection started", zap.String("id", id))
	return nil
}

// StopBypass останавливает bypass соединение
func (s *BypassService) StopBypass(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return fmt.Errorf("bypass configuration not found: %s", id)
	}

	if !s.adapter.IsRunning(id) {
		return fmt.Errorf("bypass connection not running: %s", id)
	}

	if err := s.adapter.Stop(id); err != nil {
		s.logger.Error("failed to stop bypass connection", zap.Error(err), zap.String("id", id))
		return err
	}

	config.Status = "inactive"
	config.UpdatedAt = time.Now()

	s.logger.Info("bypass connection stopped", zap.String("id", id))
	return nil
}

// GetBypassStats получает статистику bypass соединения
func (s *BypassService) GetBypassStats(ctx context.Context, id string) (*domain.BypassStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.configs[id]; !exists {
		return nil, fmt.Errorf("bypass configuration not found: %s", id)
	}

	return s.adapter.GetStats(id)
}

// DeleteBypass удаляет bypass конфигурацию
func (s *BypassService) DeleteBypass(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return fmt.Errorf("bypass configuration not found: %s", id)
	}

	// Останавливаем соединение, если запущено
	if s.adapter.IsRunning(id) {
		if err := s.adapter.Stop(id); err != nil {
			s.logger.Error("failed to stop bypass connection during deletion", zap.Error(err), zap.String("id", id))
		}
	}

	delete(s.configs, id)

	s.logger.Info("bypass configuration deleted", zap.String("id", id), zap.String("name", config.Name))
	return nil
}
