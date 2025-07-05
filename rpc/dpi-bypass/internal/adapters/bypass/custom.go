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

	// Создаем listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.LocalPort))
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Инициализируем кастомные параметры
	obfuscationMode := "hybrid" // По умолчанию гибридный режим
	chaffRatio := 0.3           // 30% мусорного трафика
	fragmentSize := 1024        // 1KB фрагменты
	timingJitter := 50          // 50ms джиттер

	// Генерируем ключ шифрования из пароля
	encryptionKey := c.generateKey(config.Password)

	conn := &customConnection{
		config:   config,
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
		stats: &domain.BypassStats{
			ID:           config.ID,
			BytesRx:      0,
			BytesTx:      0,
			Connections:  0,
			LastActivity: time.Now(),
			ErrorCount:   0,
		},
		obfuscationMode: obfuscationMode,
		chaffRatio:      chaffRatio,
		fragmentSize:    fragmentSize,
		timingJitter:    timingJitter,
		encryptionKey:   encryptionKey,
		nonceCounter:    0,
	}

	c.running[config.ID] = conn

	// Запускаем обработку соединений
	go c.handleConnections(conn)

	c.logger.Info("custom obfuscator started",
		zap.String("id", config.ID),
		zap.Int("local_port", config.LocalPort),
		zap.String("remote", config.RemoteHost),
		zap.Int("remote_port", config.RemotePort),
		zap.String("mode", obfuscationMode))

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
	conn.cancel()

	// Закрываем listener
	if err := conn.listener.Close(); err != nil {
		c.logger.Error("failed to close listener", zap.Error(err), zap.String("id", id))
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
