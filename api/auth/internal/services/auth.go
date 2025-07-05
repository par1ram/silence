package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
	"go.uber.org/zap"
)

// AuthService реализация сервиса аутентификации
type AuthService struct {
	userRepo       ports.UserRepository
	passwordHasher ports.PasswordHasher
	tokenGenerator ports.TokenGenerator
	logger         *zap.Logger
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(
	userRepo ports.UserRepository,
	passwordHasher ports.PasswordHasher,
	tokenGenerator ports.TokenGenerator,
	logger *zap.Logger,
) ports.AuthService {
	return &AuthService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		logger:         logger,
	}
}

// Register регистрирует нового пользователя
func (a *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Проверяем, существует ли пользователь
	existingUser, _ := a.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := a.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user := &domain.User{
		ID:        generateID(),
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      domain.RoleUser, // По умолчанию обычный пользователь
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем в базу
	if err := a.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := a.tokenGenerator.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info("user registered successfully", zap.String("email", user.Email))

	return &domain.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login выполняет вход пользователя
func (a *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Получаем пользователя по email
	user, err := a.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Проверяем статус пользователя
	if user.Status == domain.StatusBlocked {
		return nil, fmt.Errorf("user account is blocked")
	}

	if user.Status == domain.StatusInactive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Проверяем пароль
	if !a.passwordHasher.Verify(req.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Генерируем токен
	token, err := a.tokenGenerator.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info("user logged in successfully", zap.String("email", user.Email))

	return &domain.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// ValidateToken валидирует JWT токен
func (a *AuthService) ValidateToken(token string) (*domain.Claims, error) {
	return a.tokenGenerator.ValidateToken(token)
}

// generateID генерирует уникальный ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate random bytes for ID: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}
