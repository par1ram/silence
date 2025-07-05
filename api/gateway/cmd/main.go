package main

import (
	"github.com/par1ram/silence/api/gateway/internal/config"
	"github.com/par1ram/silence/api/gateway/internal/svc"
	"github.com/par1ram/silence/shared/app"
	"github.com/par1ram/silence/shared/container"
	"github.com/par1ram/silence/shared/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем DI контейнер
	container := container.New()

	// Создаем и регистрируем логгер
	logger := logger.NewLogger("gateway")
	container.RegisterLogger(logger)

	// Создаем контекст сервиса
	svcCtx := svc.NewServiceContext(cfg, logger)

	// Создаем приложение
	application := app.New(container)
	application.AddService(svcCtx.HTTPServer)

	// Запускаем приложение
	application.Run()
}
