package bypass

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

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
