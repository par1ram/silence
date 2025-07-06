package main

import (
	"os"
	"path/filepath"

	"github.com/par1ram/silence/api/auth/internal/adapters/database"
	"github.com/par1ram/silence/api/auth/internal/adapters/http"
	"github.com/par1ram/silence/api/auth/internal/config"
	"github.com/par1ram/silence/api/auth/internal/services"
	"github.com/par1ram/silence/shared/app"
	"github.com/par1ram/silence/shared/container"
	"github.com/par1ram/silence/shared/logger"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем DI контейнер
	container := container.New()

	// Создаем и регистрируем логгер
	logger := logger.NewLogger("auth")
	container.RegisterLogger(logger)

	// Подключаемся к базе данных
	dbConn, err := database.NewConnection(cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	// Универсальный путь к миграциям через переменную окружения
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		execPath, err := os.Executable()
		if err != nil {
			logger.Fatal("failed to get executable path", zap.Error(err))
		}
		migrationsDir = filepath.Join(filepath.Dir(execPath), "internal", "adapters", "database", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			altPath := filepath.Join("api", "auth", "internal", "adapters", "database", "migrations")
			if _, err := os.Stat(altPath); err == nil {
				logger.Info("Миграции не найдены рядом с бинарём, использую путь из исходников", zap.String("altPath", altPath))
				migrationsDir = altPath
			} else {
				logger.Fatal("Миграции не найдены ни рядом с бинарём, ни в исходниках", zap.String("tried1", migrationsDir), zap.String("tried2", altPath))
			}
		}
	}
	logger.Info("Используется путь к миграциям", zap.String("migrationsDir", migrationsDir))

	// Выполняем миграции
	migrator := database.NewMigrator(dbConn.GetDB(), logger)
	if err := migrator.RunMigrations(migrationsDir); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Создаем сервисы
	userRepo := database.NewPostgresRepository(dbConn.GetDB(), logger)
	passwordHasher := services.NewPasswordService()
	tokenGenerator := services.NewTokenService(cfg.JWTSecret, cfg.JWTExpiresIn)
	authService := services.NewAuthService(userRepo, passwordHasher, tokenGenerator, logger)
	userService := services.NewUserService(userRepo, passwordHasher)

	// Создаем HTTP обработчики
	handlers := http.NewHandlers(authService, logger)
	userHandlers := http.NewUserHandlers(userService, logger)

	// Создаем HTTP сервер
	httpServer := http.NewServer(cfg.HTTPPort, handlers, userHandlers, cfg, logger)

	// Создаем приложение
	application := app.New(container)
	application.AddService(httpServer)

	// Запускаем приложение
	application.Run()
}
