package ports

import (
	"context"

	"github.com/par1ram/silence/api/auth/internal/domain"
)

//go:generate mockgen -destination=../mocks/mocks.go -package=mocks github.com/par1ram/silence/api/auth/internal/ports AuthService,UserRepository,UserService,PasswordHasher,TokenGenerator

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *domain.UserFilter) ([]*domain.User, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error
	UpdateRole(ctx context.Context, id string, role domain.UserRole) error
}

// AuthService интерфейс для аутентификации
type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	ValidateToken(token string) (*domain.Claims, error)
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	GetMe(ctx context.Context, token string) (*domain.User, error)
}

// UserService интерфейс для управления пользователями
type UserService interface {
	CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter *domain.UserFilter) (*domain.UserListResponse, error)
	BlockUser(ctx context.Context, id string) error
	UnblockUser(ctx context.Context, id string) error
	ChangeUserRole(ctx context.Context, id string, role domain.UserRole) error
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
