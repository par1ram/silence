package main

import (
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

	// Выполняем миграции
	migrator := database.NewMigrator(dbConn.GetDB(), logger)
	if err := migrator.RunMigrations("internal/adapters/database/migrations"); err != nil {
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
