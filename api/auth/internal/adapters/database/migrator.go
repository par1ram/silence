package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// Migrator управляет миграциями базы данных
type Migrator struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewMigrator создает новый мигратор
func NewMigrator(db *sql.DB, logger *zap.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunMigrations выполняет все миграции
func (m *Migrator) RunMigrations(migrationsDir string) error {
	// Создаем таблицу для отслеживания миграций
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем список файлов миграций
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Сортируем файлы по имени
	sort.Strings(files)

	// Выполняем каждую миграцию
	for _, file := range files {
		filename := filepath.Base(file)
		if err := m.runMigration(filename, file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", filename, err)
		}
	}

	return nil
}

// createMigrationsTable создает таблицу для отслеживания миграций
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	_, err := m.db.Exec(query)
	return err
}

// runMigration выполняет одну миграцию
func (m *Migrator) runMigration(filename, filepath string) error {
	// Проверяем, была ли миграция уже выполнена
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		m.logger.Info("migration already executed", zap.String("filename", filename))
		return nil
	}

	// Читаем содержимое файла миграции
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Выполняем миграцию
	queries := strings.Split(string(content), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	// Записываем информацию о выполненной миграции
	_, err = m.db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	m.logger.Info("migration executed successfully", zap.String("filename", filename))
	return nil
}
