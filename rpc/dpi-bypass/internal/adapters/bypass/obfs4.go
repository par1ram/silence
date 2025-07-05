package bypass

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

// handleConnections обрабатывает входящие соединения
func (o *Obfs4Adapter) handleConnections(conn *obfs4Connection) {
	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			clientConn, err := conn.listener.Accept()
			if err != nil {
				if conn.ctx.Err() != nil {
					// Контекст отменен, выходим
					return
				}
				o.logger.Error("failed to accept connection", zap.Error(err), zap.String("id", conn.config.ID))
				o.incrementErrorCount(conn)
				continue
			}

			// Обрабатываем соединение в отдельной горутине
			go o.handleClientConnection(conn, clientConn)
		}
	}
}

// handleClientConnection обрабатывает клиентское соединение
func (o *Obfs4Adapter) handleClientConnection(conn *obfs4Connection, clientConn net.Conn) {
	defer clientConn.Close()

	// Увеличиваем счетчик соединений
	o.incrementConnections(conn)

	// Подключаемся к удаленному серверу
	remoteConn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", conn.config.RemoteHost, conn.config.RemotePort),
		10*time.Second)
	if err != nil {
		o.logger.Error("failed to connect to remote server",
			zap.Error(err),
			zap.String("id", conn.config.ID),
			zap.String("remote", conn.config.RemoteHost))
		o.incrementErrorCount(conn)
		return
	}
	defer remoteConn.Close()

	// Создаем каналы для передачи данных
	errChan := make(chan error, 2)

	// Копируем данные от клиента к серверу с обфускацией
	go func() {
		bytes, err := o.copyDataWithObfs(clientConn, remoteConn, conn, true)
		if err != nil {
			errChan <- err
		}
		o.updateStats(conn, bytes, 0)
	}()

	// Копируем данные от сервера к клиенту с обфускацией
	go func() {
		bytes, err := o.copyDataWithObfs(remoteConn, clientConn, conn, false)
		if err != nil {
			errChan <- err
		}
		o.updateStats(conn, 0, bytes)
	}()

	// Ждем завершения или ошибки
	select {
	case <-conn.ctx.Done():
		return
	case err := <-errChan:
		if err != nil {
			o.logger.Debug("connection error", zap.Error(err), zap.String("id", conn.config.ID))
		}
	}
}

// copyDataWithObfs копирует данные с обфускацией
func (o *Obfs4Adapter) copyDataWithObfs(src, dst net.Conn, conn *obfs4Connection, isTx bool) (int64, error) {
	buffer := make([]byte, 4096)
	var totalBytes int64

	for {
		select {
		case <-conn.ctx.Done():
			return totalBytes, nil
		default:
			// Устанавливаем таймаут для чтения
			if err := src.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				return totalBytes, err
			}

			n, err := src.Read(buffer)
			if err != nil {
				return totalBytes, err
			}

			if n > 0 {
				// Применяем обфускацию к данным
				obfuscatedData := o.obfuscateData(buffer[:n], conn)

				// Устанавливаем таймаут для записи
				if err := dst.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
					return totalBytes, err
				}

				_, err = dst.Write(obfuscatedData)
				if err != nil {
					return totalBytes, err
				}

				totalBytes += int64(n)
				o.updateLastActivity(conn)

				// Добавляем задержку IAT если включен
				if conn.iatMode {
					o.applyIATDelay(conn)
				}
			}
		}
	}
}

// obfuscateData применяет обфускацию к данным
func (o *Obfs4Adapter) obfuscateData(data []byte, conn *obfs4Connection) []byte {
	// Простая обфускация: XOR с ключом
	key := o.generateKey(conn.config.Password)
	obfuscated := make([]byte, len(data))

	for i, b := range data {
		obfuscated[i] = b ^ key[i%len(key)]
	}

	return obfuscated
}

// generateKey генерирует ключ из пароля
func (o *Obfs4Adapter) generateKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// applyIATDelay применяет задержку Inter-Arrival Time
func (o *Obfs4Adapter) applyIATDelay(conn *obfs4Connection) {
	var delay time.Duration

	switch conn.iatDist {
	case "pareto":
		// Распределение Парето
		delay = time.Duration(o.paretoDistribution(conn.iatDistMin, conn.iatDistMax)) * time.Millisecond
	case "uniform":
		// Равномерное распределение
		delay = time.Duration(o.uniformDistribution(conn.iatDistMin, conn.iatDistMax)) * time.Millisecond
	default:
		// По умолчанию небольшая случайная задержка
		delay = time.Duration(o.uniformDistribution(5, 20)) * time.Millisecond
	}

	time.Sleep(delay)
}

// paretoDistribution генерирует случайное число с распределением Парето
func (o *Obfs4Adapter) paretoDistribution(min, max int) int {
	// Упрощенная реализация распределения Парето
	u := o.randomFloat()
	// Формула Парето: x = min + (max - min) * (1 - u)^(1/alpha)
	// alpha := 2.0 // параметр формы (используется в полной формуле)
	x := float64(min) + float64(max-min)*(1-u)*(1-u)
	return int(x)
}

// uniformDistribution генерирует случайное число с равномерным распределением
func (o *Obfs4Adapter) uniformDistribution(min, max int) int {
	u := o.randomFloat()
	return min + int(u*float64(max-min))
}

// randomFloat генерирует случайное число от 0 до 1
func (o *Obfs4Adapter) randomFloat() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	var u uint64
	for i := 0; i < 8; i++ {
		u |= uint64(b[i]) << (8 * i)
	}
	return float64(u) / float64(^uint64(0))
}

// updateStats обновляет статистику
func (o *Obfs4Adapter) updateStats(conn *obfs4Connection, rx, tx int64) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.BytesRx += rx
	conn.stats.BytesTx += tx
}

// updateLastActivity обновляет время последней активности
func (o *Obfs4Adapter) updateLastActivity(conn *obfs4Connection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.LastActivity = time.Now()
}

// incrementConnections увеличивает счетчик соединений
func (o *Obfs4Adapter) incrementConnections(conn *obfs4Connection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.Connections++
}

// incrementErrorCount увеличивает счетчик ошибок
func (o *Obfs4Adapter) incrementErrorCount(conn *obfs4Connection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ErrorCount++
}
