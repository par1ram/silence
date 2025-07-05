package bypass

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"time"
)

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
