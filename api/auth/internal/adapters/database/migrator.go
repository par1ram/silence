package database

import (
	"database/sql"
	"fmt"
	"os"
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
	// Логируем параметры подключения
	if m.db != nil {
		if stats := m.db.Stats(); stats.OpenConnections == 0 {
			m.logger.Warn("DB connection pool is empty, возможно, нет соединения с БД")
		}
	}
	m.logger.Info("Запуск миграций...", zap.String("migrationsDir", migrationsDir))

	// Создаем таблицу для отслеживания миграций
	if err := m.createMigrationsTable(); err != nil {
		m.logger.Error("Ошибка создания таблицы миграций", zap.Error(err))
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем список файлов миграций
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		m.logger.Error("Ошибка чтения директории миграций", zap.Error(err))
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Сортируем файлы по имени
	sort.Strings(files)

	// Выполняем каждую миграцию
	for _, file := range files {
		filename := filepath.Base(file)
		m.logger.Info("Выполнение миграции", zap.String("filename", filename))
		if err := m.runMigration(filename, file); err != nil {
			m.logger.Error("Ошибка выполнения миграции", zap.String("filename", filename), zap.Error(err))
			return fmt.Errorf("failed to run migration %s: %w", filename, err)
		}
		m.logger.Info("Миграция успешно выполнена", zap.String("filename", filename))
	}

	m.logger.Info("Все миграции успешно применены")
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
	m.logger.Info("Проверка статуса миграции", zap.String("filename", filename))
	// Проверяем, была ли миграция уже выполнена
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		m.logger.Error("Ошибка проверки статуса миграции", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		m.logger.Info("Миграция уже была выполнена ранее", zap.String("filename", filename))
		return nil
	}

	// Читаем содержимое файла миграции
	content, err := os.ReadFile(filepath)
	if err != nil {
		m.logger.Error("Ошибка чтения файла миграции", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Выполняем миграцию
	queries := strings.Split(string(content), ";")
	for i, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		m.logger.Info("Выполнение SQL-запроса миграции", zap.String("filename", filename), zap.Int("query_index", i), zap.String("query", query))
		if _, err := m.db.Exec(query); err != nil {
			m.logger.Error("Ошибка выполнения SQL-запроса", zap.String("filename", filename), zap.Int("query_index", i), zap.Error(err))
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	// Записываем информацию о выполненной миграции
	_, err = m.db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	if err != nil {
		m.logger.Error("Ошибка записи информации о миграции", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("failed to record migration: %w", err)
	}

	m.logger.Info("Миграция успешно записана", zap.String("filename", filename))
	return nil
}
