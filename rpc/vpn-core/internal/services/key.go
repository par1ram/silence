package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"golang.org/x/crypto/curve25519"
)

// KeyGenerator реализация генератора ключей WireGuard
type KeyGenerator struct{}

// NewKeyGenerator создает новый генератор ключей
func NewKeyGenerator() ports.KeyGenerator {
	return &KeyGenerator{}
}

// GenerateKeyPair генерирует пару ключей WireGuard
func (k *KeyGenerator) GenerateKeyPair() (publicKey, privateKey string, err error) {
	// Генерируем приватный ключ
	privateKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privateKeyBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Вычисляем публичный ключ
	var publicKeyBytes [32]byte
	curve25519.ScalarBaseMult(&publicKeyBytes, (*[32]byte)(privateKeyBytes))

	// Кодируем в base64
	publicKey = base64.StdEncoding.EncodeToString(publicKeyBytes[:])
	privateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)

	return publicKey, privateKey, nil
}

// ValidatePublicKey проверяет валидность публичного ключа
func (k *KeyGenerator) ValidatePublicKey(publicKey string) bool {
	if len(publicKey) == 0 {
		return false
	}

	// Декодируем из base64
	keyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return false
	}

	// Проверяем длину (32 байта для Curve25519)
	return len(keyBytes) == 32
}
