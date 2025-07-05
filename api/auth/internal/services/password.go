package services

import (
	"fmt"

	"github.com/par1ram/silence/api/auth/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

// PasswordService реализация хеширования паролей
type PasswordService struct{}

// NewPasswordService создает новый сервис для паролей
func NewPasswordService() ports.PasswordHasher {
	return &PasswordService{}
}

// Hash хеширует пароль
func (p *PasswordService) Hash(password string) (string, error) {
	// Хешируем пароль с bcrypt (bcrypt автоматически добавляет соль)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Возвращаем хеш как строку
	return string(hash), nil
}

// Verify проверяет пароль
func (p *PasswordService) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
