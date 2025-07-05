package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/par1ram/silence/shared/container"
	"go.uber.org/zap"
)

// App базовое приложение с graceful shutdown
type App struct {
	container *container.Container
	logger    *zap.Logger
	services  []Service
}

// Service интерфейс для сервисов приложения
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// New создает новое приложение
func New(container *container.Container) *App {
	return &App{
		container: container,
		logger:    container.GetLogger(),
		services:  make([]Service, 0),
	}
}

// AddService добавляет сервис в приложение
func (a *App) AddService(service Service) {
	a.services = append(a.services, service)
}

// Run запускает приложение с graceful shutdown
func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем все сервисы
	for _, service := range a.services {
		go func(s Service) {
			a.logger.Info("starting service", zap.String("service", s.Name()))
			if err := s.Start(ctx); err != nil {
				a.logger.Error("failed to start service",
					zap.String("service", s.Name()),
					zap.Error(err))
			}
		}(service)
	}

	// Ожидаем сигнал для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	a.logger.Info("received shutdown signal, stopping services...")

	// Останавливаем все сервисы
	for _, service := range a.services {
		a.logger.Info("stopping service", zap.String("service", service.Name()))
		if err := service.Stop(ctx); err != nil {
			a.logger.Error("failed to stop service",
				zap.String("service", service.Name()),
				zap.Error(err))
		}
	}

	a.logger.Info("application stopped gracefully")
}
