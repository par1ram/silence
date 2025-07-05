package bypass

import (
	"crypto/rand"
	"crypto/sha256"
	"time"
)

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
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random bytes for randomFloat: " + err.Error())
	}
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
