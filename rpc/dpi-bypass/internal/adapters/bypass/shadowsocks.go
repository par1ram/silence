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

// ShadowsocksAdapter реализация Shadowsocks обфускации
type ShadowsocksAdapter struct {
	running map[string]*shadowsocksConnection
	mutex   sync.RWMutex
	logger  *zap.Logger
}

type shadowsocksConnection struct {
	config     *domain.BypassConfig
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *domain.BypassStats
	statsMutex sync.RWMutex
}

// NewShadowsocksAdapter создает новый Shadowsocks адаптер
func NewShadowsocksAdapter(logger *zap.Logger) *ShadowsocksAdapter {
	return &ShadowsocksAdapter{
		running: make(map[string]*shadowsocksConnection),
		logger:  logger,
	}
}

// Start запускает Shadowsocks сервер
func (s *ShadowsocksAdapter) Start(config *domain.BypassConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.running[config.ID]; exists {
		return fmt.Errorf("shadowsocks connection already running: %s", config.ID)
	}

	// Создаем контекст для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.LocalPort))
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create listener: %w", err)
	}

	conn := &shadowsocksConnection{
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
	}

	s.running[config.ID] = conn

	// Запускаем обработку соединений
	go s.handleConnections(conn)

	s.logger.Info("shadowsocks server started",
		zap.String("id", config.ID),
		zap.Int("local_port", config.LocalPort),
		zap.String("remote", config.RemoteHost),
		zap.Int("remote_port", config.RemotePort))

	return nil
}

// Stop останавливает Shadowsocks сервер
func (s *ShadowsocksAdapter) Stop(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, exists := s.running[id]
	if !exists {
		return fmt.Errorf("shadowsocks connection not found: %s", id)
	}

	// Отменяем контекст
	conn.cancel()

	// Закрываем listener
	if err := conn.listener.Close(); err != nil {
		s.logger.Error("failed to close listener", zap.Error(err), zap.String("id", id))
	}

	delete(s.running, id)

	s.logger.Info("shadowsocks server stopped", zap.String("id", id))
	return nil
}

// GetStats возвращает статистику Shadowsocks соединения
func (s *ShadowsocksAdapter) GetStats(id string) (*domain.BypassStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	conn, exists := s.running[id]
	if !exists {
		return nil, nil
	}

	conn.statsMutex.RLock()
	defer conn.statsMutex.RUnlock()

	// Возвращаем копию статистики
	stats := *conn.stats
	return &stats, nil
}

// IsRunning проверяет, запущен ли Shadowsocks сервер
func (s *ShadowsocksAdapter) IsRunning(id string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.running[id]
	return exists
}

// handleConnections обрабатывает входящие соединения
func (s *ShadowsocksAdapter) handleConnections(conn *shadowsocksConnection) {
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
				s.logger.Error("failed to accept connection", zap.Error(err), zap.String("id", conn.config.ID))
				s.incrementErrorCount(conn)
				continue
			}

			// Обрабатываем соединение в отдельной горутине
			go s.handleClientConnection(conn, clientConn)
		}
	}
}

// handleClientConnection обрабатывает клиентское соединение
func (s *ShadowsocksAdapter) handleClientConnection(conn *shadowsocksConnection, clientConn net.Conn) {
	defer clientConn.Close()

	// Увеличиваем счетчик соединений
	s.incrementConnections(conn)

	// Подключаемся к удаленному серверу
	remoteConn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", conn.config.RemoteHost, conn.config.RemotePort),
		10*time.Second)
	if err != nil {
		s.logger.Error("failed to connect to remote server",
			zap.Error(err),
			zap.String("id", conn.config.ID),
			zap.String("remote", conn.config.RemoteHost))
		s.incrementErrorCount(conn)
		return
	}
	defer remoteConn.Close()

	// Создаем каналы для передачи данных
	errChan := make(chan error, 2)

	// Копируем данные от клиента к серверу
	go func() {
		bytes, err := s.copyData(clientConn, remoteConn, conn, true)
		if err != nil {
			errChan <- err
		}
		s.updateStats(conn, bytes, 0)
	}()

	// Копируем данные от сервера к клиенту
	go func() {
		bytes, err := s.copyData(remoteConn, clientConn, conn, false)
		if err != nil {
			errChan <- err
		}
		s.updateStats(conn, 0, bytes)
	}()

	// Ждем завершения или ошибки
	select {
	case <-conn.ctx.Done():
		return
	case err := <-errChan:
		if err != nil {
			s.logger.Debug("connection error", zap.Error(err), zap.String("id", conn.config.ID))
		}
	}
}

// copyData копирует данные между соединениями
func (s *ShadowsocksAdapter) copyData(src, dst net.Conn, conn *shadowsocksConnection, isTx bool) (int64, error) {
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
				// Устанавливаем таймаут для записи
				if err := dst.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
					return totalBytes, err
				}

				_, err = dst.Write(buffer[:n])
				if err != nil {
					return totalBytes, err
				}

				totalBytes += int64(n)
				s.updateLastActivity(conn)
			}
		}
	}
}

// updateStats обновляет статистику
func (s *ShadowsocksAdapter) updateStats(conn *shadowsocksConnection, rx, tx int64) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.BytesRx += rx
	conn.stats.BytesTx += tx
}

// updateLastActivity обновляет время последней активности
func (s *ShadowsocksAdapter) updateLastActivity(conn *shadowsocksConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.LastActivity = time.Now()
}

// incrementConnections увеличивает счетчик соединений
func (s *ShadowsocksAdapter) incrementConnections(conn *shadowsocksConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.Connections++
}

// incrementErrorCount увеличивает счетчик ошибок
func (s *ShadowsocksAdapter) incrementErrorCount(conn *shadowsocksConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ErrorCount++
}
