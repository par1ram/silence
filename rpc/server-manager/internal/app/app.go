package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/par1ram/silence/rpc/server-manager/internal/adapters/database"
	"github.com/par1ram/silence/rpc/server-manager/internal/adapters/docker"
	httpadapter "github.com/par1ram/silence/rpc/server-manager/internal/adapters/http"
	"github.com/par1ram/silence/rpc/server-manager/internal/config"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"github.com/par1ram/silence/rpc/server-manager/internal/services"
	"github.com/par1ram/silence/shared/logger"
	"go.uber.org/zap"
)

// App представляет приложение Server Manager
type App struct {
	config          *config.Config
	logger          *zap.Logger
	httpServer      *httpadapter.Server
	serverService   ports.ServerService
	shutdownTimeout time.Duration
}

// resolveMigrationsDir возвращает абсолютный путь к миграциям
func resolveMigrationsDir(logger *zap.Logger) string {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		execPath, err := os.Executable()
		if err != nil {
			logger.Fatal("failed to get executable path", zap.Error(err))
		}
		migrationsDir = filepath.Join(filepath.Dir(execPath), "internal", "adapters", "database", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			altPath := filepath.Join("rpc", "server-manager", "internal", "adapters", "database", "migrations")
			if _, err := os.Stat(altPath); err == nil {
				logger.Info("Миграции не найдены рядом с бинарём, использую путь из исходников", zap.String("altPath", altPath))
				migrationsDir = altPath
			} else {
				logger.Fatal("Миграции не найдены ни рядом с бинарём, ни в исходниках", zap.String("tried1", migrationsDir), zap.String("tried2", altPath))
			}
		}
	}
	logger.Info("Используется путь к миграциям", zap.String("migrationsDir", migrationsDir))
	return migrationsDir
}

// New создает новое приложение Server Manager
func New(logger *zap.Logger) (*App, error) {
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// === АВТОМАТИЧЕСКИЕ МИГРАЦИИ ===
	migrationsDir := resolveMigrationsDir(logger)
	migrator := database.NewMigrator(db, logger)
	if err := migrator.RunMigrations(migrationsDir); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	// === END ===

	// Создаем Docker адаптер
	dockerAdapter, err := docker.NewDockerAdapter(
		cfg.Docker.Host,
		cfg.Docker.APIVersion,
		cfg.Docker.Timeout,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker adapter: %w", err)
	}

	// Создаем репозитории (заглушки для остальных репозиториев)
	serverRepo := database.NewPostgresRepository(db, logger)

	// TODO: создать остальные репозитории
	// statsRepo := database.NewStatsRepository(db, logger)
	// healthRepo := database.NewHealthRepository(db, logger)
	// scalingRepo := database.NewScalingRepository(db, logger)
	// backupRepo := database.NewBackupRepository(db, logger)
	// updateRepo := database.NewUpdateRepository(db, logger)

	// Создаем сервисы
	healthService := services.NewHealthService("server-manager", cfg.Version)
	serverService := services.NewServerService(
		serverRepo,
		nil, // statsRepo
		nil, // healthRepo
		nil, // scalingRepo
		nil, // backupRepo
		nil, // updateRepo
		dockerAdapter,
		logger,
	)

	// Создаем HTTP обработчики
	handlers := httpadapter.NewHandlers(healthService, serverService, logger)

	// Создаем HTTP сервер
	httpServer := httpadapter.NewServer(cfg.HTTPPort, handlers, logger)

	return &App{
		config:          cfg,
		logger:          logger,
		httpServer:      httpServer,
		serverService:   serverService,
		shutdownTimeout: 30 * time.Second,
	}, nil
}

// Start запускает приложение
func (a *App) Start() error {
	a.logger.Info("Starting Server Manager service",
		zap.String("http_port", a.config.HTTPPort),
		zap.String("grpc_port", a.config.GRPCPort),
		zap.String("version", a.config.Version),
	)

	// Запускаем HTTP сервер
	if err := a.httpServer.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	a.logger.Info("Server Manager service started successfully")
	return nil
}

// Shutdown останавливает приложение
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Server Manager service...")

	// Останавливаем HTTP сервер
	if err := a.httpServer.Stop(ctx); err != nil {
		a.logger.Error("Failed to stop HTTP server", zap.Error(err))
	}

	a.logger.Info("Server Manager service stopped gracefully")
	return nil
}

// ShutdownTimeout возвращает timeout для shutdown
func (a *App) ShutdownTimeout() time.Duration {
	return a.shutdownTimeout
}

// Run запускает приложение с graceful shutdown
func Run() {
	// Создаем логгер
	logger := logger.NewLogger("server-manager")
	defer func() {
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	// Создаем приложение
	app, err := New(logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Запускаем приложение
	go func() {
		if err := app.Start(); err != nil {
			logger.Fatal("Failed to start application", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Server Manager service...")

	ctx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout())
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server Manager service stopped gracefully")
}
