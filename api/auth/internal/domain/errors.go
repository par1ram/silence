package domain

import "errors"

// Кастомные типы ошибок для auth сервиса
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// IsUserAlreadyExists проверяет, является ли ошибка ошибкой существующего пользователя
func IsUserAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists)
}

// IsInvalidCredentials проверяет, является ли ошибка ошибкой неверных учетных данных
func IsInvalidCredentials(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}

// IsUserNotFound проверяет, является ли ошибка ошибкой отсутствия пользователя
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
