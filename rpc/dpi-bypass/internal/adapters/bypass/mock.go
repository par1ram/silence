package bypass

import (
	"sync"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"go.uber.org/zap"
)

// MockBypassAdapter mock реализация для тестирования
type MockBypassAdapter struct {
	running map[string]*domain.BypassConfig
	stats   map[string]*domain.BypassStats
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewMockBypassAdapter создает новый mock адаптер
func NewMockBypassAdapter(logger *zap.Logger) *MockBypassAdapter {
	return &MockBypassAdapter{
		running: make(map[string]*domain.BypassConfig),
		stats:   make(map[string]*domain.BypassStats),
		logger:  logger,
	}
}

// Start запускает mock bypass соединение
func (m *MockBypassAdapter) Start(config *domain.BypassConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running[config.ID] = config
	m.stats[config.ID] = &domain.BypassStats{
		ID:                     config.ID,
		ConfigID:               config.ID,
		SessionID:              "session_" + config.ID,
		BytesSent:              0,
		BytesReceived:          0,
		PacketsSent:            0,
		PacketsReceived:        0,
		ConnectionsEstablished: 0,
		ConnectionsFailed:      0,
		SuccessRate:            1.0,
		AverageLatency:         0,
		StartTime:              time.Now(),
		EndTime:                time.Now(),
	}

	m.logger.Info("mock: started bypass connection",
		zap.String("id", config.ID),
		zap.String("method", string(config.Method)),
		zap.String("remote", "example.com"),
		zap.Int("remote_port", 443))

	return nil
}

// Stop останавливает mock bypass соединение
func (m *MockBypassAdapter) Stop(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if config, exists := m.running[id]; exists {
		delete(m.running, id)
		m.logger.Info("mock: stopped bypass connection",
			zap.String("id", id),
			zap.String("method", string(config.Method)))
	}

	return nil
}

// GetStats возвращает статистику mock bypass соединения
func (m *MockBypassAdapter) GetStats(id string) (*domain.BypassStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if stats, exists := m.stats[id]; exists {
		// Симулируем активность
		stats.BytesReceived += 1024
		stats.BytesSent += 2048
		stats.EndTime = time.Now()
		return stats, nil
	}

	return nil, nil
}

// IsRunning проверяет, запущено ли mock bypass соединение
func (m *MockBypassAdapter) IsRunning(id string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.running[id]
	return exists
}
