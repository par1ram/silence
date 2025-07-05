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

// Obfs4Adapter реализация Obfs4 обфускации
type Obfs4Adapter struct {
	running map[string]*obfs4Connection
	mutex   sync.RWMutex
	logger  *zap.Logger
}

type obfs4Connection struct {
	config     *domain.BypassConfig
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *domain.BypassStats
	statsMutex sync.RWMutex
	// Obfs4 специфичные поля
	iatMode    bool // Inter-Arrival Time mode
	iatDist    string
	iatDistLen int
	iatDistMin int
	iatDistMax int
}

// NewObfs4Adapter создает новый Obfs4 адаптер
func NewObfs4Adapter(logger *zap.Logger) *Obfs4Adapter {
	return &Obfs4Adapter{
		running: make(map[string]*obfs4Connection),
		logger:  logger,
	}
}

// Start запускает Obfs4 сервер
func (o *Obfs4Adapter) Start(config *domain.BypassConfig) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, exists := o.running[config.ID]; exists {
		return fmt.Errorf("obfs4 connection already running: %s", config.ID)
	}

	// Создаем контекст для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.LocalPort))
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Инициализируем Obfs4 параметры
	iatMode := true // По умолчанию включен
	iatDist := "pareto"
	iatDistLen := 1
	iatDistMin := 10
	iatDistMax := 100

	conn := &obfs4Connection{
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
		iatMode:    iatMode,
		iatDist:    iatDist,
		iatDistLen: iatDistLen,
		iatDistMin: iatDistMin,
		iatDistMax: iatDistMax,
	}

	o.running[config.ID] = conn

	// Запускаем обработку соединений
	go o.handleConnections(conn)

	o.logger.Info("obfs4 server started",
		zap.String("id", config.ID),
		zap.Int("local_port", config.LocalPort),
		zap.String("remote", config.RemoteHost),
		zap.Int("remote_port", config.RemotePort),
		zap.Bool("iat_mode", iatMode))

	return nil
}

// Stop останавливает Obfs4 сервер
func (o *Obfs4Adapter) Stop(id string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	conn, exists := o.running[id]
	if !exists {
		return fmt.Errorf("obfs4 connection not found: %s", id)
	}

	// Отменяем контекст
	conn.cancel()

	// Закрываем listener
	if err := conn.listener.Close(); err != nil {
		o.logger.Error("failed to close listener", zap.Error(err), zap.String("id", id))
	}

	delete(o.running, id)

	o.logger.Info("obfs4 server stopped", zap.String("id", id))
	return nil
}

// GetStats возвращает статистику Obfs4 соединения
func (o *Obfs4Adapter) GetStats(id string) (*domain.BypassStats, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	conn, exists := o.running[id]
	if !exists {
		return nil, nil
	}

	conn.statsMutex.RLock()
	defer conn.statsMutex.RUnlock()

	// Возвращаем копию статистики
	stats := *conn.stats
	return &stats, nil
}

// IsRunning проверяет, запущен ли Obfs4 сервер
func (o *Obfs4Adapter) IsRunning(id string) bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	_, exists := o.running[id]
	return exists
}
