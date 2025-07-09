package bypass

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"go.uber.org/zap"
)

// CustomAdapter реализация кастомного обфускатора
type CustomAdapter struct {
	running map[string]*customConnection
	mutex   sync.RWMutex
	logger  *zap.Logger
}

type customConnection struct {
	config     *domain.BypassConfig
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *domain.BypassStats
	statsMutex sync.RWMutex
	// Кастомные параметры обфускации
	obfuscationMode string  // "chaff", "fragment", "timing", "hybrid"
	chaffRatio      float64 // соотношение мусорного трафика
	fragmentSize    int     // размер фрагментов
	timingJitter    int     // джиттер в миллисекундах
	encryptionKey   []byte  // ключ шифрования
	nonceCounter    uint64  // счетчик для nonce
}

// NewCustomAdapter создает новый кастомный адаптер
func NewCustomAdapter(logger *zap.Logger) *CustomAdapter {
	return &CustomAdapter{
		running: make(map[string]*customConnection),
		logger:  logger,
	}
}

// Start запускает кастомный обфускатор
func (c *CustomAdapter) Start(config *domain.BypassConfig) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.running[config.ID]; exists {
		return fmt.Errorf("custom connection already running: %s", config.ID)
	}

	// Создаем контекст для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())

	// Получаем параметры из конфигурации
	localPort := config.Parameters["local_port"]
	if localPort == "" {
		localPort = "1080"
	}

	// Создаем listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", localPort))
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create listener: %w", err)
	}

	conn := &customConnection{
		config:   config,
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
		stats: &domain.BypassStats{
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
		},
	}

	c.running[config.ID] = conn

	// Запускаем обработку соединений
	go c.handleConnections(conn)

	c.logger.Info("custom obfuscator started",
		zap.String("id", config.ID),
		zap.String("name", config.Name),
		zap.String("local_port", localPort))

	return nil
}

// Stop останавливает кастомный обфускатор
func (c *CustomAdapter) Stop(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	conn, exists := c.running[id]
	if !exists {
		return fmt.Errorf("custom connection not found: %s", id)
	}

	// Отменяем контекст
	if conn.cancel != nil {
		conn.cancel()
	}

	// Закрываем listener
	if conn.listener != nil {
		if err := conn.listener.Close(); err != nil {
			c.logger.Error("failed to close listener", zap.Error(err), zap.String("id", id))
		}
	}

	delete(c.running, id)

	c.logger.Info("custom obfuscator stopped", zap.String("id", id))
	return nil
}

// GetStats возвращает статистику кастомного соединения
func (c *CustomAdapter) GetStats(id string) (*domain.BypassStats, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	conn, exists := c.running[id]
	if !exists {
		return nil, nil
	}

	conn.statsMutex.RLock()
	defer conn.statsMutex.RUnlock()

	// Возвращаем копию статистики
	stats := *conn.stats
	return &stats, nil
}

// IsRunning проверяет, запущен ли кастомный обфускатор
func (c *CustomAdapter) IsRunning(id string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.running[id]
	return exists
}
