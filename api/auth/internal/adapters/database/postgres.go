package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
	"go.uber.org/zap"
)

// PostgresRepository реализация репозитория для PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresRepository создает новый PostgreSQL репозиторий
func NewPostgresRepository(db *sql.DB, logger *zap.Logger) ports.UserRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// Create создает нового пользователя
func (p *PostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := p.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail получает пользователя по email
func (p *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password, role, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := p.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByID получает пользователя по ID
func (p *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password, role, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update обновляет пользователя
func (p *PostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET email = $2, role = $3, status = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := p.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Role,
		user.Status,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete удаляет пользователя
func (p *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List получает список пользователей с фильтрацией
func (p *PostgresRepository) List(ctx context.Context, filter *domain.UserFilter) ([]*domain.User, int, error) {
	// Строим базовый запрос
	baseQuery := `FROM users WHERE 1=1`
	var args []interface{}
	argIndex := 1

	// Добавляем условия фильтрации
	if filter.Role != nil {
		baseQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Email != nil {
		baseQuery += fmt.Sprintf(" AND email ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	selectQuery := `SELECT id, email, password, role, status, created_at, updated_at ` + baseQuery

	// Получаем общее количество
	var total int
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Добавляем сортировку и пагинацию
	selectQuery += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		selectQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	// Выполняем запрос
	rows, err := p.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

// UpdateStatus обновляет статус пользователя
func (p *PostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error {
	query := `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateRole обновляет роль пользователя
func (p *PostgresRepository) UpdateRole(ctx context.Context, id string, role domain.UserRole) error {
	query := `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id, role)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
