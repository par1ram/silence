package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
)

// UserServiceImpl реализация UserService
type UserServiceImpl struct {
	userRepo     ports.UserRepository
	passwordHash ports.PasswordHasher
}

// NewUserService создает новый экземпляр UserService
func NewUserService(userRepo ports.UserRepository, passwordHash ports.PasswordHasher) ports.UserService {
	return &UserServiceImpl{
		userRepo:     userRepo,
		passwordHash: passwordHash,
	}
}

// CreateUser создает нового пользователя
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	// Проверяем, что email не занят
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("пользователь с email %s уже существует", req.Email)
	}

	// Хешируем пароль
	hashedPassword, err := s.passwordHash.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// Создаем пользователя
	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем в базу
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return user, nil
}

// GetUser получает пользователя по ID
func (s *UserServiceImpl) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}
	return user, nil
}

// UpdateUser обновляет пользователя
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	// Получаем текущего пользователя
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	// Обновляем поля
	if req.Email != "" {
		// Проверяем, что новый email не занят
		if req.Email != user.Email {
			existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
			if err == nil && existingUser != nil {
				return nil, fmt.Errorf("пользователь с email %s уже существует", req.Email)
			}
		}
		user.Email = req.Email
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	user.UpdatedAt = time.Now()

	// Сохраняем изменения
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return user, nil
}

// DeleteUser удаляет пользователя
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	// Проверяем, что пользователь существует
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	return s.userRepo.Delete(ctx, id)
}

// ListUsers получает список пользователей с фильтрацией
func (s *UserServiceImpl) ListUsers(ctx context.Context, filter *domain.UserFilter) (*domain.UserListResponse, error) {
	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка пользователей: %w", err)
	}

	return &domain.UserListResponse{
		Users: users,
		Total: total,
	}, nil
}

// BlockUser блокирует пользователя
func (s *UserServiceImpl) BlockUser(ctx context.Context, id string) error {
	// Проверяем, что пользователь существует
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	return s.userRepo.UpdateStatus(ctx, id, domain.StatusBlocked)
}

// UnblockUser разблокирует пользователя
func (s *UserServiceImpl) UnblockUser(ctx context.Context, id string) error {
	// Проверяем, что пользователь существует
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	return s.userRepo.UpdateStatus(ctx, id, domain.StatusActive)
}

// ChangeUserRole изменяет роль пользователя
func (s *UserServiceImpl) ChangeUserRole(ctx context.Context, id string, role domain.UserRole) error {
	// Проверяем, что пользователь существует
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	return s.userRepo.UpdateRole(ctx, id, role)
}
