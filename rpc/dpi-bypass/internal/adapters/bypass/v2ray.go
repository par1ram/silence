package bypass

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"go.uber.org/zap"
)

// V2RayAdapter реализация V2Ray обфускации
type V2RayAdapter struct {
	running map[string]*v2rayConnection
	mutex   sync.RWMutex
	logger  *zap.Logger
}

type v2rayConnection struct {
	config     *domain.BypassConfig
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *domain.BypassStats
	statsMutex sync.RWMutex
}

// NewV2RayAdapter создает новый V2Ray адаптер
func NewV2RayAdapter(logger *zap.Logger) *V2RayAdapter {
	return &V2RayAdapter{
		running: make(map[string]*v2rayConnection),
		logger:  logger,
	}
}

// Start запускает V2Ray сервер
func (v *V2RayAdapter) Start(config *domain.BypassConfig) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if _, exists := v.running[config.ID]; exists {
		return fmt.Errorf("v2ray connection already running: %s", config.ID)
	}

	// Создаем контекст для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем listener с TLS для WebSocket
	var listener net.Listener
	var err error

	if config.Encryption == "tls" {
		// Создаем самоподписанный сертификат для тестирования
		cert, err := generateSelfSignedCert()
		if err != nil {
			cancel()
			return fmt.Errorf("failed to generate certificate: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		listener, err = tls.Listen("tcp", fmt.Sprintf(":%d", config.LocalPort), tlsConfig)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create TLS listener: %w", err)
		}
	} else {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", config.LocalPort))
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create TCP listener: %w", err)
		}
	}

	conn := &v2rayConnection{
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

	v.running[config.ID] = conn

	// Запускаем обработку соединений
	go v.handleConnections(conn)

	v.logger.Info("v2ray server started",
		zap.String("id", config.ID),
		zap.Int("local_port", config.LocalPort),
		zap.String("remote", config.RemoteHost),
		zap.Int("remote_port", config.RemotePort),
		zap.String("encryption", config.Encryption))

	return nil
}

// Stop останавливает V2Ray сервер
func (v *V2RayAdapter) Stop(id string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	conn, exists := v.running[id]
	if !exists {
		return fmt.Errorf("v2ray connection not found: %s", id)
	}

	// Отменяем контекст
	conn.cancel()

	// Закрываем listener
	if err := conn.listener.Close(); err != nil {
		v.logger.Error("failed to close listener", zap.Error(err), zap.String("id", id))
	}

	delete(v.running, id)

	v.logger.Info("v2ray server stopped", zap.String("id", id))
	return nil
}

// GetStats возвращает статистику V2Ray соединения
func (v *V2RayAdapter) GetStats(id string) (*domain.BypassStats, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	conn, exists := v.running[id]
	if !exists {
		return nil, nil
	}

	conn.statsMutex.RLock()
	defer conn.statsMutex.RUnlock()

	// Возвращаем копию статистики
	stats := *conn.stats
	return &stats, nil
}

// IsRunning проверяет, запущен ли V2Ray сервер
func (v *V2RayAdapter) IsRunning(id string) bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	_, exists := v.running[id]
	return exists
}

// handleConnections обрабатывает входящие соединения
func (v *V2RayAdapter) handleConnections(conn *v2rayConnection) {
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
				v.logger.Error("failed to accept connection", zap.Error(err), zap.String("id", conn.config.ID))
				v.incrementErrorCount(conn)
				continue
			}

			// Обрабатываем соединение в отдельной горутине
			go v.handleClientConnection(conn, clientConn)
		}
	}
}

// handleClientConnection обрабатывает клиентское соединение
func (v *V2RayAdapter) handleClientConnection(conn *v2rayConnection, clientConn net.Conn) {
	defer clientConn.Close()

	// Увеличиваем счетчик соединений
	v.incrementConnections(conn)

	// Подключаемся к удаленному серверу
	remoteConn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", conn.config.RemoteHost, conn.config.RemotePort),
		10*time.Second)
	if err != nil {
		v.logger.Error("failed to connect to remote server",
			zap.Error(err),
			zap.String("id", conn.config.ID),
			zap.String("remote", conn.config.RemoteHost))
		v.incrementErrorCount(conn)
		return
	}
	defer remoteConn.Close()

	// Создаем каналы для передачи данных
	errChan := make(chan error, 2)

	// Копируем данные от клиента к серверу
	go func() {
		bytes, err := v.copyData(clientConn, remoteConn, conn, true)
		if err != nil {
			errChan <- err
		}
		v.updateStats(conn, bytes, 0)
	}()

	// Копируем данные от сервера к клиенту
	go func() {
		bytes, err := v.copyData(remoteConn, clientConn, conn, false)
		if err != nil {
			errChan <- err
		}
		v.updateStats(conn, 0, bytes)
	}()

	// Ждем завершения или ошибки
	select {
	case <-conn.ctx.Done():
		return
	case err := <-errChan:
		if err != nil {
			v.logger.Debug("connection error", zap.Error(err), zap.String("id", conn.config.ID))
		}
	}
}

// copyData копирует данные между соединениями
func (v *V2RayAdapter) copyData(src, dst net.Conn, conn *v2rayConnection, isTx bool) (int64, error) {
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
				v.updateLastActivity(conn)
			}
		}
	}
}

// updateStats обновляет статистику
func (v *V2RayAdapter) updateStats(conn *v2rayConnection, rx, tx int64) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.BytesRx += rx
	conn.stats.BytesTx += tx
}

// updateLastActivity обновляет время последней активности
func (v *V2RayAdapter) updateLastActivity(conn *v2rayConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.LastActivity = time.Now()
}

// incrementConnections увеличивает счетчик соединений
func (v *V2RayAdapter) incrementConnections(conn *v2rayConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.Connections++
}

// incrementErrorCount увеличивает счетчик ошибок
func (v *V2RayAdapter) incrementErrorCount(conn *v2rayConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ErrorCount++
}
