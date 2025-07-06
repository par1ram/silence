package services

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyGenerator(t *testing.T) {
	t.Run("создание генератора ключей", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		assert.NotNil(t, keyGen)
		assert.IsType(t, &KeyGenerator{}, keyGen)
	})

	t.Run("генерация пары ключей", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		publicKey, privateKey, err := keyGen.GenerateKeyPair()

		assert.NoError(t, err)
		assert.NotEmpty(t, publicKey)
		assert.NotEmpty(t, privateKey)
		assert.NotEqual(t, publicKey, privateKey)

		// Проверяем, что ключи в base64 формате
		_, err = base64.StdEncoding.DecodeString(publicKey)
		assert.NoError(t, err)

		_, err = base64.StdEncoding.DecodeString(privateKey)
		assert.NoError(t, err)
	})

	t.Run("множественная генерация ключей", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Генерируем несколько пар ключей
		keys := make(map[string]bool)
		for i := 0; i < 10; i++ {
			publicKey, privateKey, err := keyGen.GenerateKeyPair()
			assert.NoError(t, err)

			// Проверяем уникальность
			assert.False(t, keys[publicKey], "публичный ключ должен быть уникальным")
			assert.False(t, keys[privateKey], "приватный ключ должен быть уникальным")

			keys[publicKey] = true
			keys[privateKey] = true
		}
	})

	t.Run("валидация корректного публичного ключа", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		publicKey, _, err := keyGen.GenerateKeyPair()
		assert.NoError(t, err)

		isValid := keyGen.ValidatePublicKey(publicKey)
		assert.True(t, isValid)
	})

	t.Run("валидация некорректного публичного ключа", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		testCases := []struct {
			name     string
			key      string
			expected bool
		}{
			{
				name:     "пустой ключ",
				key:      "",
				expected: false,
			},
			{
				name:     "некорректный base64",
				key:      "invalid-base64-key!@#",
				expected: false,
			},
			{
				name:     "слишком короткий ключ",
				key:      base64.StdEncoding.EncodeToString([]byte("short")),
				expected: false,
			},
			{
				name:     "слишком длинный ключ",
				key:      base64.StdEncoding.EncodeToString(make([]byte, 64)),
				expected: false,
			},
			{
				name:     "ключ с пробелами",
				key:      " " + base64.StdEncoding.EncodeToString(make([]byte, 32)) + " ",
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				isValid := keyGen.ValidatePublicKey(tc.key)
				assert.Equal(t, tc.expected, isValid)
			})
		}
	})

	t.Run("валидация ключа правильной длины", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ правильной длины (32 байта)
		correctKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
		isValid := keyGen.ValidatePublicKey(correctKey)
		assert.True(t, isValid)
	})

	t.Run("валидация ключа с специальными символами", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ с различными байтами
		keyBytes := make([]byte, 32)
		for i := range keyBytes {
			keyBytes[i] = byte(i)
		}
		key := base64.StdEncoding.EncodeToString(keyBytes)

		isValid := keyGen.ValidatePublicKey(key)
		assert.True(t, isValid)
	})

	t.Run("валидация ключа с нулевыми байтами", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ из нулевых байтов
		zeroKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
		isValid := keyGen.ValidatePublicKey(zeroKey)
		assert.True(t, isValid)
	})

	t.Run("валидация ключа с padding", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ, который может иметь padding
		keyBytes := make([]byte, 32)
		keyBytes[31] = 0 // Последний байт 0 может создать padding
		key := base64.StdEncoding.EncodeToString(keyBytes)

		isValid := keyGen.ValidatePublicKey(key)
		assert.True(t, isValid)
	})

	t.Run("валидация ключа с различными символами base64", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ с различными байтами, которые дают разные символы base64
		keyBytes := []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		}
		key := base64.StdEncoding.EncodeToString(keyBytes)

		isValid := keyGen.ValidatePublicKey(key)
		assert.True(t, isValid)
	})

	t.Run("валидация ключа с пробелами и переносами строк", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		correctKey := base64.StdEncoding.EncodeToString(make([]byte, 32))

		// Добавляем пробелы и переносы строк
		keyWithSpaces := " " + correctKey + " "
		keyWithNewlines := "\n" + correctKey + "\n"
		keyWithTabs := "\t" + correctKey + "\t"

		assert.False(t, keyGen.ValidatePublicKey(keyWithSpaces))
		assert.True(t, keyGen.ValidatePublicKey(keyWithNewlines))
		assert.False(t, keyGen.ValidatePublicKey(keyWithTabs))
	})

	t.Run("валидация ключа с URL-безопасным base64", func(t *testing.T) {
		keyGen := NewKeyGenerator()

		// Создаем ключ с URL-безопасным base64
		keyBytes := make([]byte, 32)
		urlSafeKey := base64.URLEncoding.EncodeToString(keyBytes)

		// URL-безопасный base64 может быть валидным, если он корректно декодируется
		isValid := keyGen.ValidatePublicKey(urlSafeKey)
		assert.True(t, isValid)
	})
}
