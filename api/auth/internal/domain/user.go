package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserRole роль пользователя
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
)

// UserStatus статус пользователя
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBlocked  UserStatus = "blocked"
)

// User пользователь системы
type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // Не отправляем в JSON
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Claims JWT claims
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse ответ с токеном
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// CreateUserRequest запрос на создание пользователя
type CreateUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

// UpdateUserRequest запрос на обновление пользователя
type UpdateUserRequest struct {
	Email  string     `json:"email,omitempty"`
	Role   UserRole   `json:"role,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// UserListResponse ответ со списком пользователей
type UserListResponse struct {
	Users []*User `json:"users"`
	Total int     `json:"total"`
}

// UserFilter фильтр для поиска пользователей
type UserFilter struct {
	Role   *UserRole   `json:"role,omitempty"`
	Status *UserStatus `json:"status,omitempty"`
	Email  *string     `json:"email,omitempty"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
