package ports

import (
	"context"

	"github.com/par1ram/silence/api/auth/internal/domain"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// AuthService интерфейс для аутентификации
type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	ValidateToken(token string) (*domain.Claims, error)
}

// PasswordHasher интерфейс для хеширования паролей
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// TokenGenerator интерфейс для генерации JWT токенов
type TokenGenerator interface {
	GenerateToken(user *domain.User) (string, error)
	ValidateToken(token string) (*domain.Claims, error)
}
