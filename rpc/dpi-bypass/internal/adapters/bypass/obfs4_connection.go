package bypass

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

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
