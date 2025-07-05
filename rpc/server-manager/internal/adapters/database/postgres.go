package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
)

// PostgresRepository реализация репозитория для PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresRepository создает новый PostgreSQL репозиторий
func NewPostgresRepository(db *sql.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// Create создает новый сервер
func (r *PostgresRepository) Create(ctx context.Context, server *domain.Server) error {
	query := `
		INSERT INTO servers (id, name, type, status, region, ip, port, cpu, memory, disk, network, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	server.ID = uuid.New().String()
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		server.ID, server.Name, server.Type, server.Status, server.Region,
		server.IP, server.Port, server.CPU, server.Memory, server.Disk, server.Network,
		server.CreatedAt, server.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	r.logger.Info("server created", zap.String("id", server.ID), zap.String("name", server.Name))
	return nil
}

// GetByID получает сервер по ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Server, error) {
	query := `
		SELECT id, name, type, status, region, ip, port, cpu, memory, disk, network, created_at, updated_at, deleted_at
		FROM servers WHERE id = $1 AND deleted_at IS NULL
	`

	server := &domain.Server{}
	var ip sql.NullString
	var port sql.NullInt32
	var cpu sql.NullFloat64
	var memory sql.NullFloat64
	var disk sql.NullFloat64
	var network sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID, &server.Name, &server.Type, &server.Status, &server.Region,
		&ip, &port, &cpu, &memory, &disk, &network,
		&server.CreatedAt, &server.UpdatedAt, &server.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// Обрабатываем NULL значения
	if ip.Valid {
		server.IP = ip.String
	}
	if port.Valid {
		server.Port = int(port.Int32)
	}
	if cpu.Valid {
		server.CPU = cpu.Float64
	}
	if memory.Valid {
		server.Memory = memory.Float64
	}
	if disk.Valid {
		server.Disk = disk.Float64
	}
	if network.Valid {
		server.Network = network.Float64
	}

	return server, nil
}

// List получает список серверов с фильтрами
func (r *PostgresRepository) List(ctx context.Context, filters map[string]interface{}) ([]*domain.Server, error) {
	query := `
		SELECT id, name, type, status, region, ip, port, cpu, memory, disk, network, created_at, updated_at, deleted_at
		FROM servers WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	// Добавляем фильтры
	if serverType, ok := filters["type"].(domain.ServerType); ok {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, serverType)
		argIndex++
	}

	if region, ok := filters["region"].(string); ok {
		query += fmt.Sprintf(" AND region = $%d", argIndex)
		args = append(args, region)
		argIndex++
	}

	if status, ok := filters["status"].(domain.ServerStatus); ok {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	defer rows.Close()

	var servers []*domain.Server
	for rows.Next() {
		server := &domain.Server{}
		var ip sql.NullString
		var port sql.NullInt32
		var cpu sql.NullFloat64
		var memory sql.NullFloat64
		var disk sql.NullFloat64
		var network sql.NullFloat64

		err := rows.Scan(
			&server.ID, &server.Name, &server.Type, &server.Status, &server.Region,
			&ip, &port, &cpu, &memory, &disk, &network,
			&server.CreatedAt, &server.UpdatedAt, &server.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}

		// Обрабатываем NULL значения
		if ip.Valid {
			server.IP = ip.String
		}
		if port.Valid {
			server.Port = int(port.Int32)
		}
		if cpu.Valid {
			server.CPU = cpu.Float64
		}
		if memory.Valid {
			server.Memory = memory.Float64
		}
		if disk.Valid {
			server.Disk = disk.Float64
		}
		if network.Valid {
			server.Network = network.Float64
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// Update обновляет сервер
func (r *PostgresRepository) Update(ctx context.Context, server *domain.Server) error {
	query := `
		UPDATE servers 
		SET name = $2, type = $3, status = $4, region = $5, ip = $6, port = $7, 
		    cpu = $8, memory = $9, disk = $10, network = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	server.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		server.ID, server.Name, server.Type, server.Status, server.Region,
		server.IP, server.Port, server.CPU, server.Memory, server.Disk, server.Network,
		server.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", server.ID)
	}

	r.logger.Info("server updated", zap.String("id", server.ID))
	return nil
}

// Delete удаляет сервер (soft delete)
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE servers SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", id)
	}

	r.logger.Info("server deleted", zap.String("id", id))
	return nil
}

// GetByType получает серверы по типу
func (r *PostgresRepository) GetByType(ctx context.Context, serverType domain.ServerType) ([]*domain.Server, error) {
	return r.List(ctx, map[string]interface{}{"type": serverType})
}

// GetByRegion получает серверы по региону
func (r *PostgresRepository) GetByRegion(ctx context.Context, region string) ([]*domain.Server, error) {
	return r.List(ctx, map[string]interface{}{"region": region})
}

// GetByStatus получает серверы по статусу
func (r *PostgresRepository) GetByStatus(ctx context.Context, status domain.ServerStatus) ([]*domain.Server, error) {
	return r.List(ctx, map[string]interface{}{"status": status})
}
