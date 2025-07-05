package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/par1ram/silence/api/auth/internal/config"
	"go.uber.org/zap"
)

// Connection управляет подключением к базе данных
type Connection struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewConnection создает новое подключение к базе данных
func NewConnection(cfg *config.Config, logger *zap.Logger) (*Connection, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение с retry
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			logger.Warn("failed to ping database, retrying...",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.String("error", err.Error()))

			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
			}

			// Ждем перед следующей попыткой
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		break
	}

	logger.Info("connected to database",
		zap.String("host", cfg.DBHost),
		zap.String("port", cfg.DBPort),
		zap.String("database", cfg.DBName))

	return &Connection{
		db:     db,
		logger: logger,
	}, nil
}

// GetDB возвращает подключение к базе данных
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Close закрывает подключение к базе данных
func (c *Connection) Close() error {
	c.logger.Info("closing database connection")
	return c.db.Close()
}
