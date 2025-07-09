package bypass

import (
	"time"
)

// updateStats обновляет статистику
func (c *CustomAdapter) updateStats(conn *customConnection, rx, tx int64) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.BytesReceived += rx
	conn.stats.BytesSent += tx
}

// updateLastActivity обновляет время последней активности
func (c *CustomAdapter) updateLastActivity(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.EndTime = time.Now()
}

// incrementConnections увеличивает счетчик соединений
func (c *CustomAdapter) incrementConnections(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ConnectionsEstablished++
}

// incrementErrorCount увеличивает счетчик ошибок
func (c *CustomAdapter) incrementErrorCount(conn *customConnection) {
	conn.statsMutex.Lock()
	defer conn.statsMutex.Unlock()

	conn.stats.ConnectionsFailed++
}
