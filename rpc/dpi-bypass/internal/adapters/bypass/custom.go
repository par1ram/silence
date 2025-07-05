package bypass

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

// handleConnections обрабатывает входящие соединения
func (c *CustomAdapter) handleConnections(conn *customConnection) {
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
				c.logger.Error("failed to accept connection", zap.Error(err), zap.String("id", conn.config.ID))
				c.incrementErrorCount(conn)
				continue
			}

			// Обрабатываем соединение в отдельной горутине
			go c.handleClientConnection(conn, clientConn)
		}
	}
}

// handleClientConnection обрабатывает клиентское соединение
func (c *CustomAdapter) handleClientConnection(conn *customConnection, clientConn net.Conn) {
	defer clientConn.Close()

	// Увеличиваем счетчик соединений
	c.incrementConnections(conn)

	// Подключаемся к удаленному серверу
	remoteConn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", conn.config.RemoteHost, conn.config.RemotePort),
		10*time.Second)
	if err != nil {
		c.logger.Error("failed to connect to remote server",
			zap.Error(err),
			zap.String("id", conn.config.ID),
			zap.String("remote", conn.config.RemoteHost))
		c.incrementErrorCount(conn)
		return
	}
	defer remoteConn.Close()

	// Создаем каналы для передачи данных
	errChan := make(chan error, 2)

	// Копируем данные от клиента к серверу с кастомной обфускацией
	go func() {
		bytes, err := c.copyDataWithCustomObfs(clientConn, remoteConn, conn, true)
		if err != nil {
			errChan <- err
		}
		c.updateStats(conn, bytes, 0)
	}()

	// Копируем данные от сервера к клиенту с кастомной обфускацией
	go func() {
		bytes, err := c.copyDataWithCustomObfs(remoteConn, clientConn, conn, false)
		if err != nil {
			errChan <- err
		}
		c.updateStats(conn, 0, bytes)
	}()

	// Ждем завершения или ошибки
	select {
	case <-conn.ctx.Done():
		return
	case err := <-errChan:
		if err != nil {
			c.logger.Debug("connection error", zap.Error(err), zap.String("id", conn.config.ID))
		}
	}
}

// copyDataWithCustomObfs копирует данные с кастомной обфускацией
func (c *CustomAdapter) copyDataWithCustomObfs(src, dst net.Conn, conn *customConnection, isTx bool) (int64, error) {
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
				// Применяем кастомную обфускацию
				obfuscatedData, err := c.applyCustomObfuscation(buffer[:n], conn)
				if err != nil {
					return totalBytes, err
				}

				// Устанавливаем таймаут для записи
				if err := dst.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
					return totalBytes, err
				}

				_, err = dst.Write(obfuscatedData)
				if err != nil {
					return totalBytes, err
				}

				totalBytes += int64(n)
				c.updateLastActivity(conn)

				// Применяем кастомные задержки
				c.applyCustomTiming(conn)
			}
		}
	}
}

// applyCustomObfuscation применяет кастомную обфускацию
func (c *CustomAdapter) applyCustomObfuscation(data []byte, conn *customConnection) ([]byte, error) {
	switch conn.obfuscationMode {
	case "chaff":
		return c.applyChaffObfuscation(data, conn)
	case "fragment":
		return c.applyFragmentObfuscation(data, conn)
	case "hybrid":
		return c.applyHybridObfuscation(data, conn)
	default:
		return c.applyHybridObfuscation(data, conn)
	}
}

// applyChaffObfuscation добавляет мусорный трафик
func (c *CustomAdapter) applyChaffObfuscation(data []byte, conn *customConnection) ([]byte, error) {
	// Шифруем данные
	encryptedData, err := c.encryptData(data, conn)
	if err != nil {
		return nil, err
	}

	// Добавляем мусорный трафик
	chaffData := c.generateChaffData(len(encryptedData), conn)

	// Перемешиваем реальные данные с мусором
	result := make([]byte, 0, len(encryptedData)+len(chaffData))

	// Добавляем заголовок с информацией о размерах
	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[0:4], uint32(len(encryptedData)))
	binary.BigEndian.PutUint32(header[4:8], uint32(len(chaffData)))
	result = append(result, header...)

	// Добавляем реальные данные
	result = append(result, encryptedData...)

	// Добавляем мусорные данные
	result = append(result, chaffData...)

	return result, nil
}

// applyFragmentObfuscation разбивает данные на фрагменты
func (c *CustomAdapter) applyFragmentObfuscation(data []byte, conn *customConnection) ([]byte, error) {
	// Шифруем данные
	encryptedData, err := c.encryptData(data, conn)
	if err != nil {
		return nil, err
	}

	// Разбиваем на фрагменты
	fragments := c.fragmentData(encryptedData, conn.fragmentSize)

	// Собираем результат с заголовками фрагментов
	result := make([]byte, 0, len(encryptedData)+len(fragments)*8)

	for i, fragment := range fragments {
		// Заголовок фрагмента: [номер][размер][данные]
		header := make([]byte, 8)
		binary.BigEndian.PutUint32(header[0:4], uint32(i))
		binary.BigEndian.PutUint32(header[4:8], uint32(len(fragment)))
		result = append(result, header...)
		result = append(result, fragment...)
	}

	return result, nil
}

// applyHybridObfuscation комбинирует несколько методов
func (c *CustomAdapter) applyHybridObfuscation(data []byte, conn *customConnection) ([]byte, error) {
	// Сначала фрагментируем
	fragmented, err := c.applyFragmentObfuscation(data, conn)
	if err != nil {
		return nil, err
	}

	// Затем добавляем мусорный трафик
	return c.applyChaffObfuscation(fragmented, conn)
}

// generateChaffData генерирует мусорные данные
func (c *CustomAdapter) generateChaffData(realDataSize int, conn *customConnection) []byte {
	// Вычисляем размер мусорных данных на основе соотношения
	chaffSize := int(float64(realDataSize) * conn.chaffRatio)
	if chaffSize < 64 {
		chaffSize = 64 // Минимальный размер
	}

	chaffData := make([]byte, chaffSize)

	// Генерируем псевдослучайные данные
	for i := 0; i < chaffSize; i++ {
		chaffData[i] = byte(conn.nonceCounter % 256)
		conn.nonceCounter++
	}

	return chaffData
}

// fragmentData разбивает данные на фрагменты
func (c *CustomAdapter) fragmentData(data []byte, fragmentSize int) [][]byte {
	var fragments [][]byte

	for i := 0; i < len(data); i += fragmentSize {
		end := i + fragmentSize
		if end > len(data) {
			end = len(data)
		}
		fragments = append(fragments, data[i:end])
	}

	return fragments
}

// encryptData шифрует данные
func (c *CustomAdapter) encryptData(data []byte, conn *customConnection) ([]byte, error) {
	block, err := aes.NewCipher(conn.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Генерируем nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Шифруем данные
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// generateKey генерирует ключ из пароля
func (c *CustomAdapter) generateKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// applyCustomTiming применяет кастомные задержки
func (c *CustomAdapter) applyCustomTiming(conn *customConnection) {
	if conn.timingJitter > 0 {
		// Используем синусоидальную функцию для имитации человеческого поведения
		timeOffset := time.Now().UnixNano() / int64(time.Millisecond)
		jitter := int(math.Sin(float64(timeOffset)/1000.0) * float64(conn.timingJitter))
		if jitter < 0 {
			jitter = -jitter
		}
		if jitter > 0 {
			time.Sleep(time.Duration(jitter) * time.Millisecond)
		}
	}
}

// updateStats обновляет статистику
func (c *CustomAdapter) updateStats(conn *customConnection, rx, tx int64) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.BytesRx += rx
	conn.stats.BytesTx += tx
}

// updateLastActivity обновляет время последней активности
func (c *CustomAdapter) updateLastActivity(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.LastActivity = time.Now()
}

// incrementConnections увеличивает счетчик соединений
func (c *CustomAdapter) incrementConnections(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.Connections++
}

// incrementErrorCount увеличивает счетчик ошибок
func (c *CustomAdapter) incrementErrorCount(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ErrorCount++
}
