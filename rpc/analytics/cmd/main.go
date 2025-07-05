package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/par1ram/silence/rpc/analytics/internal/app"
	"github.com/par1ram/silence/shared/logger"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	logger := logger.NewLogger("analytics")

	// Создание приложения
	app, err := app.New(logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.String("error", err.Error()))
	}

	// Запуск приложения
	go func() {
		if err := app.Start(); err != nil {
			logger.Fatal("Failed to start application", zap.String("error", err.Error()))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down analytics service...")

	ctx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout())
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Analytics service stopped gracefully")
}
